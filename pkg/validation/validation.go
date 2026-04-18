package validation

import "strings"

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

func (e *Errors) HasAny() bool {
	return len(e.items) > 0
}

func (e *Errors) Details() map[string]any {
	if len(e.items) == 0 {
		return nil
	}
	return map[string]any{"fields": e.items}
}
