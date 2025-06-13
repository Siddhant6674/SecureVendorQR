package user

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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
	router.HandleFunc("/vendorInfo/{phone}", h.handleVendorInfo).Methods("GET")
	router.HandleFunc("/Register", h.handleRegister).Methods("POST")
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

	//if vendor doesn't then we register it
	err = h.store.CreateVendor(types.Vendor{
		FirstName: Payload.FirstName,
		LastName:  Payload.LastName,
		Phone:     Payload.Phone,
		PanNO:     Payload.PanNO,
		AdharNo:   Payload.AdharNo,
		GSTno:     Payload.GSTno,
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

// function to handle the get request after scanning the QR
func (h *Handler) handleVendorInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	phone := vars["phone"]

	utils.SendOTP(phone)

}

func (h *Handler) handleVerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req types.OTPrequest

	if err := utils.ParseJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
	}

	if !utils.ValidateOTP(req.Phone, req.OTP) {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid or expired otp"))
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "OTP verified successfully"})
}
