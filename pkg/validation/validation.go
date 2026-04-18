package validation

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type FieldError struct {
	Field   string `json:"field"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type Errors struct {
	items []FieldError
}

func New() *Errors {
	return &Errors{items: make([]FieldError, 0)}
}

func (e *Errors) Required(field string, value string, message string) {
	if strings.TrimSpace(value) != "" {
		return
	}
	e.items = append(e.items, FieldError{Field: field, Reason: "required", Message: message})
}

func (e *Errors) MinLength(field string, value string, min int, message string) {
	if len(strings.TrimSpace(value)) >= min {
		return
	}
	e.items = append(e.items, FieldError{Field: field, Reason: "min_length", Message: message})
}

func (e *Errors) UUID(field string, value string, message string) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return
	}
	if _, err := uuid.Parse(trimmedValue); err == nil {
		return
	}
	e.items = append(e.items, FieldError{Field: field, Reason: "invalid_uuid", Message: message})
}

func (e *Errors) RFC3339(field string, value string, message string) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return
	}
	if _, err := time.Parse(time.RFC3339, trimmedValue); err == nil {
		return
	}
	e.items = append(e.items, FieldError{Field: field, Reason: "invalid_datetime", Message: message})
}

func (e *Errors) HasAny() bool {
	return len(e.items) > 0
}

func (e *Errors) Details() map[string]any {
	if len(e.items) == 0 {
		return nil
	}
	return map[string]any{"fields": e.items}
}
