package apierror

import "encoding/json"

type messagePayload struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func messageWithCode(code string, message string, details map[string]any) string {
	payload, err := json.Marshal(messagePayload{
		Code:    code,
		Message: message,
		Details: details,
	})
	if err != nil {
		return message
	}

	return string(payload)
}

func parseMessage(raw string) (string, string, map[string]any) {
	var payload messagePayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return "", raw, nil
	}

	return payload.Code, payload.Message, payload.Details
}
