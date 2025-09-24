package utils

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
)

// GenerateQRCode generates a QR code for a ticket
func GenerateQRCode(ticketID uint, eventID uint, userID uint) (string, error) {
	// Create unique QR data using UUID and timestamp
	qrData := fmt.Sprintf("TICKET-%d-%d-%d-%s-%d",
		ticketID, eventID, userID, uuid.New().String(), time.Now().UnixNano())

	// Generate QR code as bytes
	qrBytes, err := qrcode.Encode(qrData, qrcode.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code: %v", err)
	}

	return string(qrBytes), nil
}

// ValidateQRCode validates QR code data
func ValidateQRCode(qrData string) (bool, error) {
	// Basic validation - check if QR data follows expected format
	expectedPrefix := "TICKET-"
	if len(qrData) < len(expectedPrefix) {
		return false, fmt.Errorf("invalid QR code format")
	}

	return true, nil
}