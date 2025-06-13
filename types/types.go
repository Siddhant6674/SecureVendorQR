package types

import "time"

type VendorStore interface {
	GetVendorByPhone(phone string) (*Vendor, error)
	CreateVendor(Vendor) error
}
type Vendor struct {
	ID        int       `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Phone     string    `json:"phone"`
	PanNO     string    `json:"panNO"`
	AdharNo   string    `json:"adharNo"`
	GSTno     string    `json:"gstNo"`
	CreatedAt time.Time `json:"createdAt"`
}

type RegisterVendorPayload struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Phone     string `json:"phone" validate:"required,len=10,numeric"`
	PanNO     string `json:"panNO" validate:"required,len=10,alphanum"`
	AdharNo   string `json:"adharNo" validate:"required,len=12,numeric"`
	GSTno     string `json:"gstNo" validate:"required"`
}

type OTPrequest struct {
	Phone string `json:"phone" validate:"required.len=10,numeric"`
	OTP   string `json:"otp"`
}

type Fast2SMSPayload struct {
	Route    string `json:"route"`
	SenderID string `json:"sender_id"`
	Message  string `json:"message"`
	Language string `json:"language"`
	Flash    string `json:"flash"`
	Numbers  string `json:"numbers"` // comma-separated
}
