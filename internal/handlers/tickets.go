package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"event-ticketing-system/internal/models"
	"event-ticketing-system/pkg/utils"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)


// TicketHandler handles ticket related requests
type TicketHandler struct {
	db *gorm.DB
}

// NewTicketHandler creates a new ticket handler
func NewTicketHandler(db *gorm.DB) *TicketHandler {
	return &TicketHandler{db: db}
}

// PurchaseTicketRequest represents the purchase ticket request payload
type PurchaseTicketRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1,max=10"`
}

// GetTickets retrieves tickets for the current user or all tickets (admin)
func (h *TicketHandler) GetTickets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Context().Value("user_id")
	if userID == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not authenticated"})
		return
	}

	userRole := r.Context().Value("user_role")

	var tickets []models.Ticket

	if userRole == "admin" {
		// Admin can see all tickets
		if err := h.db.Preload("Event").Preload("User").Preload("AttendanceLogs").Find(&tickets).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve tickets"})
			return
		}
	} else {
		// Regular users can only see their own tickets
		if err := h.db.Preload("Event").Preload("AttendanceLogs").Where("user_id = ?", userID).Find(&tickets).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve tickets"})
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tickets)
}

// GetTicket retrieves a specific ticket by ID
func (h *TicketHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get ID from URL parameters (Gorilla Mux way)
	vars := mux.Vars(r)
	id := vars["id"]
	ticketID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid ticket ID"})
		return
	}

	userID := r.Context().Value("user_id")
	if userID == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not authenticated"})
		return
	}

	userRole := r.Context().Value("user_role")

	var ticket models.Ticket
	query := h.db.Preload("Event").Preload("User").Preload("AttendanceLogs")

	if userRole == "admin" {
		// Admin can see any ticket
		query = query.Where("id = ?", ticketID)
	} else {
		// Regular users can only see their own tickets
		query = query.Where("id = ? AND user_id = ?", ticketID, userID)
	}

	if err := query.First(&ticket).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ticket not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve ticket"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ticket)
}

// PurchaseTicket handles ticket purchase for an event
func (h *TicketHandler) PurchaseTicket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get ID from URL parameters (Gorilla Mux way)
	vars := mux.Vars(r)
	eventID := vars["id"]
	eventIDUint, err := strconv.ParseUint(eventID, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid event ID"})
		return
	}

	userID := r.Context().Value("user_id")
	if userID == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not authenticated"})
		return
	}

	var req PurchaseTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Check if event exists
	var event models.Event
	if err := h.db.Where("id = ?", eventIDUint).First(&event).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Event not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve event"})
		return
	}

	// Check if event date is in the future
	if event.Date.Before(time.Now()) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Cannot purchase tickets for past events"})
		return
	}

	// Check available capacity
	var existingTicketsCount int64
	h.db.Model(&models.Ticket{}).Where("event_id = ?", eventIDUint).Count(&existingTicketsCount)
	availableCapacity := event.Capacity - int(existingTicketsCount)

	if req.Quantity > availableCapacity {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not enough tickets available"})
		return
	}

	// Generate tickets
	var tickets []models.Ticket
	for i := 0; i < req.Quantity; i++ {
		// Generate unique QR code using utility function
		qrCode, err := utils.GenerateQRCode(uint(eventIDUint), userID.(uint), uint(i+1))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to generate QR code"})
			return
		}

		ticket := models.Ticket{
			EventID: uint(eventIDUint),
			UserID:  userID.(uint),
			QRCode:  qrCode,
			Status:  "valid",
		}

		if err := h.db.Create(&ticket).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create ticket"})
			return
		}

		tickets = append(tickets, ticket)
	}

	response := map[string]interface{}{
		"message": "Tickets purchased successfully",
		"tickets": tickets,
		"total":   len(tickets),
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// ValidateTicket validates a ticket using QR code (admin only)
func (h *TicketHandler) ValidateTicket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get ID from URL parameters (Gorilla Mux way)
	vars := mux.Vars(r)
	id := vars["id"]
	ticketID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid ticket ID"})
		return
	}

	var ticket models.Ticket
	if err := h.db.Where("id = ?", ticketID).First(&ticket).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ticket not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve ticket"})
		return
	}

	// Check if ticket is already used
	if ticket.Status == "used" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ticket has already been used"})
		return
	}

	// Mark ticket as used and create attendance log
	ticket.Status = "used"
	if err := h.db.Save(&ticket).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to validate ticket"})
		return
	}

	// Create attendance log
	attendanceLog := models.AttendanceLog{
		TicketID:    ticket.ID,
		CheckedInAt: time.Now(),
	}

	if err := h.db.Create(&attendanceLog).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create attendance log"})
		return
	}

	response := map[string]interface{}{
		"message": "Ticket validated successfully",
		"ticket":  ticket,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetEventAttendees retrieves attendees for a specific event (admin only)
func (h *TicketHandler) GetEventAttendees(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get ID from URL parameters (Gorilla Mux way)
	vars := mux.Vars(r)
	eventID := vars["id"]
	eventIDUint, err := strconv.ParseUint(eventID, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid event ID"})
		return
	}

	var tickets []models.Ticket
	if err := h.db.Preload("User").Preload("AttendanceLogs").Where("event_id = ?", eventIDUint).Find(&tickets).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve attendees"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tickets)
}

// ExportAttendees exports attendees for a specific event as CSV (admin only)
func (h *TicketHandler) ExportAttendees(w http.ResponseWriter, r *http.Request) {
	// Get ID from URL parameters (Gorilla Mux way)
	vars := mux.Vars(r)
	eventID := vars["id"]
	eventIDUint, err := strconv.ParseUint(eventID, 10, 32)
	if err != nil {
		http.Error(w, `{"error": "Invalid event ID"}`, http.StatusBadRequest)
		return
	}

	var tickets []models.Ticket
	if err := h.db.Preload("User").Preload("AttendanceLogs").Where("event_id = ?", eventIDUint).Find(&tickets).Error; err != nil {
		http.Error(w, `{"error": "Failed to retrieve attendees"}`, http.StatusInternalServerError)
		return
	}

	// Set CSV headers
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=attendees_event_%s.csv", eventID))

	// Create CSV writer
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write CSV header
	writer.Write([]string{"Ticket ID", "User Name", "User Email", "Status", "Checked In At", "Purchase Date"})

	// Write attendee data
	for _, ticket := range tickets {
		checkedInAt := ""
		if len(ticket.AttendanceLogs) > 0 {
			checkedInAt = ticket.AttendanceLogs[0].CheckedInAt.Format("2006-01-02 15:04:05")
		}

		writer.Write([]string{
			fmt.Sprintf("%d", ticket.ID),
			ticket.User.Name,
			ticket.User.Email,
			ticket.Status,
			checkedInAt,
			ticket.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
}