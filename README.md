#SecureVendorQR

This project is a Go-based backend system for vendor registration and secure QR-based access using OTP (One-Time Password) verification. It is designed to securely verify vendor identity before granting access to their information via QR code.

Features:-

📝 Vendor Registration: Vendors can register using basic information.

📤 QR Code Generation: A QR code is generated upon registration containing the vendor’s access link.

🔐 OTP Verification: On scanning the QR, an OTP is sent to the registered phone number for verification.

📦 Redis Integration: Stores OTP with an expiry (2 minutes) and tracks verification status.

🧪 OTP Validation Endpoint: Checks OTP correctness and marks phone number as verified.

⚙️ Modular Code Architecture: Separates concerns via handlers, utils, and types.

Tech Stack:-

Go (Golang) – Backend server

Redis – Temporary OTP and status storage

Gorilla Mux – HTTP routing

Fast2SMS / JSON Mode – Placeholder for SMS API integration

MySQL – Vendor database

Project Status
✅ OTP Generation and Redis Integration

✅ API for sending and verifying OTP

✅ QR Code generation and vendor registration

⚠️ SMS sending via Fast2SMS is not active due to paid requirement. OTP is currently returned as JSON for testing.
