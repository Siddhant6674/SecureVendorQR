package user

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Siddhant6674/vendorQr/types"
)

type store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *store {
	return &store{db: db}
}

// creating fucntion for scanning rows into database
func scanRowsIntoVendor(rows *sql.Rows) (*types.Vendor, error) {
	vendor := new(types.Vendor)

	err := rows.Scan(
		&vendor.ID,
		&vendor.FirstName,
		&vendor.LastName,
		&vendor.Phone,
		&vendor.PanNO,
		&vendor.AdharNo,
		&vendor.GSTno,
		&vendor.CreatedAt,
		&vendor.Password,
	)
	if err != nil {
		return nil, err
	}
	return vendor, nil
}

// Get vendor by phone number from scaning into database
func (s *store) GetVendorByPhone(phone string) (*types.Vendor, error) {
	// Run a SQL query to get all columns of the vendor where phone matches
	rows, err := s.db.Query("SELECT*FROM vendor WHERE phone= ?", phone)
	if err != nil {
		// If the query fails, return the error
		return nil, err
	}

	// Create a new empty Vendor struct to store data
	u := new(types.Vendor)
	// Loop through the rows returned by the query
	for rows.Next() {
		// Scan the row into the Vendor struct
		u, err = scanRowsIntoVendor(rows)
		if err != nil {
			// If scanning fails, return the error
			return nil, err
		}
	}
	// If no vendor was found (ID still zero), return "user not found" error
	//If this false then vendor exist so return u,nil (no error)
	if u.ID == 0 {
		return nil, fmt.Errorf("user not found")
	}
	// Return the found vendor and no error
	return u, nil
}

// resgistering vendor to database

func (s *store) CreateVendor(vendor types.Vendor) error {
	_, err := s.db.Exec("INSERT INTO vendor(firstname,lastname,phone,panNo,adharNo,gstNo,password)VALUES(?,?,?,?,?,?,?)",
		vendor.FirstName,
		vendor.LastName,
		vendor.Phone,
		vendor.PanNO,
		vendor.AdharNo,
		vendor.GSTno,
		vendor.Password)
	if err != nil {
		log.Println("Insert error:", err)
		return err
	}
	return nil
}

func (s *store) GetVendorByID(ID int) (*types.Vendor, error) {
	rows, err := s.db.Query("SELECT * FROM vendor WHERE ID=?", ID)
	if err != nil {
		return nil, err
	}

	u := new(types.Vendor)
	for rows.Next() {
		u, err = scanRowsIntoVendor(rows)
		if err != nil {
			return nil, err
		}
	}

	if u.ID == 0 {
		return nil, fmt.Errorf("user not found")
	}
	return u, nil
}
