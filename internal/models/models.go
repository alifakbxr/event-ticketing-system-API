package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/jinzhu/gorm"
)

// User represents a user in the system
type User struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	Name      string    `json:"name" gorm:"not null" validate:"required"`
	Email     string    `json:"email" gorm:"unique;not null" validate:"required,email"`
	Password  string    `json:"-" gorm:"not null" validate:"required"`
	Role      string    `json:"role" gorm:"default:'user'" validate:"required,oneof=admin user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Event represents an event in the system
type Event struct {
	ID          uint      `json:"id" gorm:"primary_key"`
	Title       string    `json:"title" gorm:"not null" validate:"required"`
	Description string    `json:"description" gorm:"not null" validate:"required"`
	Date        time.Time `json:"date" gorm:"not null" validate:"required"`
	Location    string    `json:"location" gorm:"not null" validate:"required"`
	Capacity    int       `json:"capacity" gorm:"not null" validate:"required,min=1"`
	Price       float64   `json:"price" gorm:"not null" validate:"required,min=0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Tickets []Ticket `json:"tickets,omitempty" gorm:"foreignkey:EventID"`
}

// Ticket represents a ticket for an event
type Ticket struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	EventID   uint      `json:"event_id" gorm:"not null"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	QRCode    string    `json:"qr_code" gorm:"unique;not null"`
	Status    string    `json:"status" gorm:"default:'valid'" validate:"required,oneof=valid used"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Event           Event            `json:"event,omitempty" gorm:"foreignkey:EventID"`
	User            User             `json:"user,omitempty" gorm:"foreignkey:UserID"`
	AttendanceLogs  []AttendanceLog  `json:"attendance_logs,omitempty" gorm:"foreignkey:TicketID"`
}

// AttendanceLog represents a check-in record for a ticket
type AttendanceLog struct {
	ID           uint      `json:"id" gorm:"primary_key"`
	TicketID     uint      `json:"ticket_id" gorm:"not null"`
	CheckedInAt  time.Time `json:"checked_in_at" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Ticket Ticket `json:"ticket,omitempty" gorm:"foreignkey:TicketID"`
}

// TableName overrides the table name used by User to `users`
func (User) TableName() string {
	return "users"
}

// TableName overrides the table name used by Event to `events`
func (Event) TableName() string {
	return "events"
}

// TableName overrides the table name used by Ticket to `tickets`
func (Ticket) TableName() string {
	return "tickets"
}

// TableName overrides the table name used by AttendanceLog to `attendance_logs`
func (AttendanceLog) TableName() string {
	return "attendance_logs"
}

// BeforeCreate hook to hash password before saving
func (u *User) BeforeCreate(scope *gorm.Scope) error {
	if len(u.Password) == 0 {
		return nil
	}

	hashedPassword, err := hashPassword(u.Password)
	if err != nil {
		return err
	}

	return scope.SetColumn("Password", hashedPassword)
}

// BeforeUpdate hook to hash password before updating
func (u *User) BeforeUpdate(scope *gorm.Scope) error {
	if len(u.Password) == 0 {
		return nil
	}

	hashedPassword, err := hashPassword(u.Password)
	if err != nil {
		return err
	}

	return scope.SetColumn("Password", hashedPassword)
}

// hashPassword hashes the password using bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}