package user

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Siddhant6674/vendorQr/config"
	"github.com/Siddhant6674/vendorQr/service/auth"
	"github.com/Siddhant6674/vendorQr/types"
	"github.com/Siddhant6674/vendorQr/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	store types.VendorStore
}

func Newhandler(store types.VendorStore) *Handler {
	return &Handler{store: store}
}

// All register routes with its method
func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/login", h.handleLogin).Methods("POST")
	router.HandleFunc("/vendorInfo/{phone}", h.handleVendorInfo).Methods("GET")
	router.HandleFunc("/Register", (h.handleRegister)).Methods("POST")
	router.HandleFunc("/accessQR", auth.WithJWT(h.handleAccessQR, h.store)).Methods("POST")
	router.HandleFunc("/accessinformation", auth.WithJWT(h.handleAccessInformation, h.store)).Methods("POST")
	router.HandleFunc("/recoverQR/{phone}", auth.WithJWT(h.handleRecoverQR, h.store)).Methods("GET")
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var payload types.LoginUserPayload

	//get user payload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
	}

	//validate user paylaod
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	//check user exist or not
	u, err := h.store.GetVendorByPhone(payload.Phone)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("not found, invalid phone number or password"))
	}

	//compared password
	if !auth.ComparedPasswords(u.Password, []byte(payload.Password)) {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid phone number or password"))
		return
	}

	//create JWT token
	secret := []byte(config.Envs.JWTSecret)
	token, err := auth.CreateJWT(secret, u.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"Token": token})
}

// Handler for registering vendor
func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var Payload types.RegisterVendorPayload

	//get vendor payload
	if err := utils.ParseJSON(r, &Payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
	}

	//validate vendor payload
	if err := utils.Validate.Struct(Payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	// check if vendor exist or not by phone no
	_, err := h.store.GetVendorByPhone(Payload.Phone)
	log.Printf("GetVendorByPhone error: %v", err)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user with phone %s is already exist", Payload.Phone))
	}

	hashedPassword, err := auth.HashPassword(Payload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
	}

	//if vendor doesn't then we register it
	err = h.store.CreateVendor(types.Vendor{
		FirstName: Payload.FirstName,
		LastName:  Payload.LastName,
		Phone:     Payload.Phone,
		PanNO:     Payload.PanNO,
		AdharNo:   Payload.AdharNo,
		GSTno:     Payload.GSTno,
		Password:  hashedPassword,
	})
	// Try to create vendor in DB
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("duplicate value error: %v", err))
			return // Stop here, don't generate QR
		}
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to create vendor: %v", err))
		return
	}
	// If no error, proceed with QR code generation...

	// Generate QR code for new vendor
	url := fmt.Sprintf("http://localhost:8080/api/v1/vendorInfo/%s", Payload.Phone)
	filePath := "vendorQr/vendor_qr/" + Payload.Phone + "_qrcode.png"

	err = os.MkdirAll("vendorQr/vendor_qr", os.ModePerm)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to create folder: %v", err))
		return
	}

	qrcode, err := utils.GenrateQrCode(url, filePath)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to generate qr code:%v", err))
	}

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusCreated)
	w.Write(qrcode)

}

// handler of get request which send otp in response for reteriving info by used of accessinformation handler
// which is post request
func (h *Handler) handleVendorInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	phone := vars["phone"]

	otp, err := utils.SendOTP(phone)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to generate otp"))
		return
	}
	utils.WriteJSON(w, http.StatusOK, otp)

}

// endpoint of get req for on demand qrcode which send otp in response,
// and otp used in endpoint accessQR which is post req
func (h *Handler) handleRecoverQR(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	phone := vars["phone"]

	otp, err := utils.SendOTP(phone)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to generate otp"))
		return
	}
	utils.WriteJSON(w, http.StatusOK, otp)
}

// endpoint of post request which validate the otp and handover the QR code on demand
func (h *Handler) handleAccessQR(w http.ResponseWriter, r *http.Request) {
	var req types.OTPrequest

	if err := utils.ParseJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
	}

	if !utils.ValidateOTP(req.Phone, req.OTP) {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid or expired otp"))
		return
	}

	// Mark OTP as verified
	err := utils.MarkOTPVerified(req.Phone)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "OTP verified successfully"})

	err = utils.ServeQRcode(w, req.Phone)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
}

// endpoint of post req which validate otp and handover the vendor info
func (h *Handler) handleAccessInformation(w http.ResponseWriter, r *http.Request) {
	var req types.OTPrequest

	if err := utils.ParseJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
	}

	if !utils.ValidateOTP(req.Phone, req.OTP) {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid or expired otp"))
		return
	}

	err := utils.MarkOTPVerified(req.Phone)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
	}

	vendor, err := h.store.GetVendorByPhone(req.Phone)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, err)
	}

	utils.WriteJSON(w, http.StatusOK, vendor)
}
