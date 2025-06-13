package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"

	"time"

	"github.com/Siddhant6674/vendorQr/config"
	"github.com/Siddhant6674/vendorQr/types"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/skip2/go-qrcode"
)

var Validate = validator.New()
var rdb *redis.Client
var ctx = context.Background()
var fast2SMSAPIKey string // replace this

// function to intialize redis server
func InitRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // redis server addressed
	})
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Error connecting to redis:", err)
	}
	log.Println("Redis connected successfully")
}

func SendSMSFast2SMS(phone string, message string) error {
	fast2SMSAPIKey = config.Envs.APIkey
	payload := types.Fast2SMSPayload{
		Route:    "v3",
		SenderID: "Sender",
		Message:  message,
		Language: "english",
		Flash:    "0",
		Numbers:  phone,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", "https://www.fast2sms.com/dev/bulkV2", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("authorization", fast2SMSAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	log.Println("Fast2SMS Response Body:", string(body))

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 response from Fast2SMS: %d", res.StatusCode)
	}

	return nil
}

// function to generate OTP
func GenerateOTP(length int) string {
	const digits = "0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // Local generator

	otp := make([]byte, length)
	for i := range otp {
		otp[i] = digits[r.Intn(len(digits))]
	}
	return string(otp)
}

// function to store otp in redis
func StoreOTP(phone, OTP string) error {
	err := rdb.Set(ctx, phone, OTP, 2*time.Minute).Err() // OTP expired in 2 minutes
	if err != nil {
		return fmt.Errorf("failed to store otp in redis: %v ", err)
	}
	return nil
}

// fucntion to send OTP
func SendOTP(Phone string) error {
	otp := GenerateOTP(6)

	//Store otp in redis
	StoreOTP(Phone, otp)

	//Create message
	message := fmt.Sprintf("Your OTP for QR code access is %s", otp)
	fmt.Printf("DEBUG OTP for %s: %s\n", Phone, otp)
	//send via FastSMS
	err := SendSMSFast2SMS(Phone, message)
	if err != nil {
		return fmt.Errorf("failed to send sms %v", err)
	}
	return nil
}

// ValidateOTP checks if the OTP is correct and hasn't expired
func ValidateOTP(phone, otp string) bool {
	storedOTP, err := rdb.Get(ctx, phone).Result()
	if err != nil {
		return false // Return false if OTP not found or expired
	}
	return storedOTP == otp
}

// function to generate QR code
func GenrateQrCode(url, filePath string) ([]byte, error) {
	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		return nil, err
	}
	// Save file to disk
	err = os.WriteFile(filePath, png, 0644)
	if err != nil {
		return nil, err
	}
	// Return PNG bytes
	return png, nil
}

// fucntion for converting request body data into json form
func ParseJSON(r *http.Request, Payload any) error {
	if r.Body == nil {
		return fmt.Errorf("missing reuest body")
	}
	return json.NewDecoder(r.Body).Decode(Payload)
}

// function for encode the data/msg into json format
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// function for wrting error
func WriteError(w http.ResponseWriter, status int, err error) {
	WriteJSON(w, status, map[string]string{"error": err.Error()})
}
