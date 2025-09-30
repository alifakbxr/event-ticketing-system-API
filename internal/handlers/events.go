package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"event-ticketing-system/internal/models"

	"github.com/gorilla/mux"
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
func (h *EventHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var events []models.Event
	if err := h.db.Preload("Tickets").Find(&events).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve events"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(events)
}

// GetEvent retrieves a specific event by ID
func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get ID from URL parameters (Gorilla Mux way)
	vars := mux.Vars(r)
	id := vars["id"]
	eventID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid event ID"})
		return
	}

	var event models.Event
	if err := h.db.Preload("Tickets").Where("id = ?", eventID).First(&event).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Event not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve event"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(event)
}

// CreateEvent creates a new event (admin only)
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
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
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create event"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(event)
}

// UpdateEvent updates an existing event (admin only)
func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get ID from URL parameters (Gorilla Mux way)
	vars := mux.Vars(r)
	id := vars["id"]
	eventID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid event ID"})
		return
	}

	var event models.Event
	if err := h.db.Where("id = ?", eventID).First(&event).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Event not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve event"})
		return
	}

	var req UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
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
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update event"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(event)
}

// DeleteEvent deletes an event (admin only)
func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get ID from URL parameters (Gorilla Mux way)
	vars := mux.Vars(r)
	id := vars["id"]
	eventID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid event ID"})
		return
	}

	// Check if event exists
	var event models.Event
	if err := h.db.Where("id = ?", eventID).First(&event).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Event not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve event"})
		return
	}

	// Check if there are any tickets for this event
	var ticketCount int64
	h.db.Model(&models.Ticket{}).Where("event_id = ?", eventID).Count(&ticketCount)
	if ticketCount > 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Cannot delete event with existing tickets"})
		return
	}

	if err := h.db.Delete(&event).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete event"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Event deleted successfully"})
}