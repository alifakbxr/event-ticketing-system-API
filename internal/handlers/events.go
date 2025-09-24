package handlers

import (
	"net/http"
	"strconv"
	"time"

	"event-ticketing-system/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)


// EventHandler handles event related requests
type EventHandler struct {
	db *gorm.DB
}

// NewEventHandler creates a new event handler
func NewEventHandler(db *gorm.DB) *EventHandler {
	return &EventHandler{db: db}
}

// CreateEventRequest represents the create event request payload
type CreateEventRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description" binding:"required"`
	Date        time.Time `json:"date" binding:"required"`
	Location    string    `json:"location" binding:"required"`
	Capacity    int       `json:"capacity" binding:"required,min=1"`
	Price       float64   `json:"price" binding:"required,min=0"`
}

// UpdateEventRequest represents the update event request payload
type UpdateEventRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Location    string    `json:"location"`
	Capacity    int       `json:"capacity"`
	Price       float64   `json:"price"`
}

// GetEvents retrieves all events
// swagger:operation GET /api/events events getEvents
// ---
// summary: Get all events
// description: Retrieves a list of all available events
// tags:
// - Events
// responses:
//   200:
//     description: List of events retrieved successfully
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/Event"
//   500:
//     description: Internal server error
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
func (h *EventHandler) GetEvents(c *gin.Context) {
	var events []models.Event
	if err := h.db.Preload("Tickets").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve events"})
		return
	}

	c.JSON(http.StatusOK, events)
}

// GetEvent retrieves a specific event by ID
// swagger:operation GET /api/events/{id} events getEvent
// ---
// summary: Get event by ID
// description: Retrieves a specific event by its ID
// tags:
// - Events
// parameters:
// - name: id
//   in: path
//   description: Event ID
//   required: true
//   type: integer
//   format: int64
// responses:
//   200:
//     description: Event retrieved successfully
//     schema:
//       "$ref": "#/definitions/Event"
//   400:
//     description: Invalid event ID
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
func (h *EventHandler) GetEvent(c *gin.Context) {
	id := c.Param("id")
	eventID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var event models.Event
	if err := h.db.Preload("Tickets").Where("id = ?", eventID).First(&event).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// CreateEvent creates a new event (admin only)
// swagger:operation POST /api/events events createEvent
// ---
// summary: Create a new event
// description: Creates a new event (admin only)
// tags:
// - Events
// security:
// - Bearer: []
// parameters:
// - name: request
//   in: body
//   description: Event creation data
//   required: true
//   schema:
//     "$ref": "#/definitions/CreateEventRequest"
// responses:
//   201:
//     description: Event created successfully
//     schema:
//       "$ref": "#/definitions/Event"
//   400:
//     description: Invalid request data
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
func (h *EventHandler) CreateEvent(c *gin.Context) {
	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event := models.Event{
		Title:       req.Title,
		Description: req.Description,
		Date:        req.Date,
		Location:    req.Location,
		Capacity:    req.Capacity,
		Price:       req.Price,
	}

	if err := h.db.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, event)
}

// UpdateEvent updates an existing event (admin only)
// swagger:operation PUT /api/events/{id} events updateEvent
// ---
// summary: Update an existing event
// description: Updates an existing event by ID (admin only)
// tags:
// - Events
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
//   description: Event update data
//   required: true
//   schema:
//     "$ref": "#/definitions/UpdateEventRequest"
// responses:
//   200:
//     description: Event updated successfully
//     schema:
//       "$ref": "#/definitions/Event"
//   400:
//     description: Invalid request data or event ID
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
//     description: Event not found
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   500:
//     description: Internal server error
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
func (h *EventHandler) UpdateEvent(c *gin.Context) {
	id := c.Param("id")
	eventID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var event models.Event
	if err := h.db.Where("id = ?", eventID).First(&event).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
		return
	}

	var req UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.Title != "" {
		event.Title = req.Title
	}
	if req.Description != "" {
		event.Description = req.Description
	}
	if !req.Date.IsZero() {
		event.Date = req.Date
	}
	if req.Location != "" {
		event.Location = req.Location
	}
	if req.Capacity > 0 {
		event.Capacity = req.Capacity
	}
	if req.Price >= 0 {
		event.Price = req.Price
	}

	if err := h.db.Save(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// DeleteEvent deletes an event (admin only)
// swagger:operation DELETE /api/events/{id} events deleteEvent
// ---
// summary: Delete an event
// description: Deletes an existing event by ID (admin only). Cannot delete events with existing tickets.
// tags:
// - Events
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
//     description: Event deleted successfully
//     schema:
//       type: object
//       properties:
//         message:
//           type: string
//           example: "Event deleted successfully"
//   400:
//     description: Invalid event ID or event has existing tickets
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
//     description: Event not found
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   500:
//     description: Internal server error
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
func (h *EventHandler) DeleteEvent(c *gin.Context) {
	id := c.Param("id")
	eventID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Check if event exists
	var event models.Event
	if err := h.db.Where("id = ?", eventID).First(&event).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
		return
	}

	// Check if there are any tickets for this event
	var ticketCount int64
	h.db.Model(&models.Ticket{}).Where("event_id = ?", eventID).Count(&ticketCount)
	if ticketCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete event with existing tickets"})
		return
	}

	if err := h.db.Delete(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}