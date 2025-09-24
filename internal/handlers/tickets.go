package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"event-ticketing-system/internal/models"
	"event-ticketing-system/pkg/utils"

	"github.com/gin-gonic/gin"
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
// swagger:operation GET /api/tickets tickets getTickets
// ---
// summary: Get tickets
// description: Retrieves tickets for the current user or all tickets (admin only)
// tags:
// - Tickets
// security:
// - Bearer: []
// responses:
//   200:
//     description: Tickets retrieved successfully
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/Ticket"
//   401:
//     description: Unauthorized
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   500:
//     description: Internal server error
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
func (h *TicketHandler) GetTickets(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userRole, _ := c.Get("user_role")

	var tickets []models.Ticket

	if userRole == "admin" {
		// Admin can see all tickets
		if err := h.db.Preload("Event").Preload("User").Preload("AttendanceLogs").Find(&tickets).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tickets"})
			return
		}
	} else {
		// Regular users can only see their own tickets
		if err := h.db.Preload("Event").Preload("AttendanceLogs").Where("user_id = ?", userID).Find(&tickets).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tickets"})
			return
		}
	}

	c.JSON(http.StatusOK, tickets)
}

// GetTicket retrieves a specific ticket by ID
// swagger:operation GET /api/tickets/{id} tickets getTicket
// ---
// summary: Get ticket by ID
// description: Retrieves a specific ticket by ID (users can only see their own tickets)
// tags:
// - Tickets
// security:
// - Bearer: []
// parameters:
// - name: id
//   in: path
//   description: Ticket ID
//   required: true
//   type: integer
//   format: int64
// responses:
//   200:
//     description: Ticket retrieved successfully
//     schema:
//       "$ref": "#/definitions/Ticket"
//   400:
//     description: Invalid ticket ID
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   401:
//     description: Unauthorized
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   404:
//     description: Ticket not found
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   500:
//     description: Internal server error
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
func (h *TicketHandler) GetTicket(c *gin.Context) {
	id := c.Param("id")
	ticketID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userRole, _ := c.Get("user_role")

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
			c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve ticket"})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

// PurchaseTicket handles ticket purchase for an event
// swagger:operation POST /api/events/{id}/purchase tickets purchaseTicket
// ---
// summary: Purchase tickets for an event
// description: Purchases tickets for a specific event
// tags:
// - Tickets
// security:
// - Bearer: []
// parameters:
// - name: id
//   in: path
//   description: Event ID
//   required: true
//   type: integer
//   format: int64
// - name: request
//   in: body
//   description: Ticket purchase data
//   required: true
//   schema:
//     "$ref": "#/definitions/PurchaseTicketRequest"
// responses:
//   201:
//     description: Tickets purchased successfully
//     schema:
//       type: object
//       properties:
//         message:
//           type: string
//           example: "Tickets purchased successfully"
//         tickets:
//           type: array
//           items:
//             "$ref": "#/definitions/Ticket"
//         total:
//           type: integer
//           example: 2
//   400:
//     description: Invalid request data, event ID, or insufficient capacity
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   401:
//     description: Unauthorized
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   404:
//     description: Event not found
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   500:
//     description: Internal server error
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
func (h *TicketHandler) PurchaseTicket(c *gin.Context) {
	eventID := c.Param("id")
	eventIDUint, err := strconv.ParseUint(eventID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req PurchaseTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if event exists
	var event models.Event
	if err := h.db.Where("id = ?", eventIDUint).First(&event).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
		return
	}

	// Check if event date is in the future
	if event.Date.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot purchase tickets for past events"})
		return
	}

	// Check available capacity
	var existingTicketsCount int64
	h.db.Model(&models.Ticket{}).Where("event_id = ?", eventIDUint).Count(&existingTicketsCount)
	availableCapacity := event.Capacity - int(existingTicketsCount)

	if req.Quantity > availableCapacity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough tickets available"})
		return
	}

	// Generate tickets
	var tickets []models.Ticket
	for i := 0; i < req.Quantity; i++ {
		// Generate unique QR code using utility function
		qrCode, err := utils.GenerateQRCode(uint(eventIDUint), userID.(uint), uint(i+1))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate QR code"})
			return
		}

		ticket := models.Ticket{
			EventID: uint(eventIDUint),
			UserID:  userID.(uint),
			QRCode:  qrCode,
			Status:  "valid",
		}

		if err := h.db.Create(&ticket).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ticket"})
			return
		}

		tickets = append(tickets, ticket)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tickets purchased successfully",
		"tickets": tickets,
		"total":   len(tickets),
	})
}

// ValidateTicket validates a ticket using QR code (admin only)
// swagger:operation POST /api/tickets/{id}/validate tickets validateTicket
// ---
// summary: Validate a ticket
// description: Validates a ticket and marks it as used (admin only)
// tags:
// - Tickets
// security:
// - Bearer: []
// parameters:
// - name: id
//   in: path
//   description: Ticket ID
//   required: true
//   type: integer
//   format: int64
// responses:
//   200:
//     description: Ticket validated successfully
//     schema:
//       "$ref": "#/definitions/TicketValidationResponse"
//   400:
//     description: Invalid ticket ID or ticket already used
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   401:
//     description: Unauthorized
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   403:
//     description: Admin access required
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   404:
//     description: Ticket not found
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   500:
//     description: Internal server error
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
func (h *TicketHandler) ValidateTicket(c *gin.Context) {
	id := c.Param("id")
	ticketID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	var ticket models.Ticket
	if err := h.db.Where("id = ?", ticketID).First(&ticket).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve ticket"})
		return
	}

	// Check if ticket is already used
	if ticket.Status == "used" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ticket has already been used"})
		return
	}

	// Mark ticket as used and create attendance log
	ticket.Status = "used"
	if err := h.db.Save(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate ticket"})
		return
	}

	// Create attendance log
	attendanceLog := models.AttendanceLog{
		TicketID:    ticket.ID,
		CheckedInAt: time.Now(),
	}

	if err := h.db.Create(&attendanceLog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create attendance log"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ticket validated successfully",
		"ticket":  ticket,
	})
}

// GetEventAttendees retrieves attendees for a specific event (admin only)
// swagger:operation GET /api/events/{id}/attendees tickets getEventAttendees
// ---
// summary: Get event attendees
// description: Retrieves all attendees for a specific event (admin only)
// tags:
// - Tickets
// security:
// - Bearer: []
// parameters:
// - name: id
//   in: path
//   description: Event ID
//   required: true
//   type: integer
//   format: int64
// responses:
//   200:
//     description: Attendees retrieved successfully
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/Ticket"
//   400:
//     description: Invalid event ID
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   401:
//     description: Unauthorized
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   403:
//     description: Admin access required
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   500:
//     description: Internal server error
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
func (h *TicketHandler) GetEventAttendees(c *gin.Context) {
	eventID := c.Param("id")
	eventIDUint, err := strconv.ParseUint(eventID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var tickets []models.Ticket
	if err := h.db.Preload("User").Preload("AttendanceLogs").Where("event_id = ?", eventIDUint).Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve attendees"})
		return
	}

	c.JSON(http.StatusOK, tickets)
}

// ExportAttendees exports attendees for a specific event as CSV (admin only)
// swagger:operation GET /api/events/{id}/attendees/export tickets exportAttendees
// ---
// summary: Export event attendees
// description: Exports attendees for a specific event as CSV file (admin only)
// tags:
// - Tickets
// security:
// - Bearer: []
// parameters:
// - name: id
//   in: path
//   description: Event ID
//   required: true
//   type: integer
//   format: int64
// produces:
// - text/csv
// responses:
//   200:
//     description: CSV file downloaded successfully
//     schema:
//       type: file
//   400:
//     description: Invalid event ID
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   401:
//     description: Unauthorized
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   403:
//     description: Admin access required
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   500:
//     description: Internal server error
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
func (h *TicketHandler) ExportAttendees(c *gin.Context) {
	eventID := c.Param("id")
	eventIDUint, err := strconv.ParseUint(eventID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var tickets []models.Ticket
	if err := h.db.Preload("User").Preload("AttendanceLogs").Where("event_id = ?", eventIDUint).Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve attendees"})
		return
	}

	// Set CSV headers
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment;filename=attendees_event_%s.csv", eventID))

	// Create CSV writer
	writer := csv.NewWriter(c.Writer)
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