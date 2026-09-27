package main

import (
	"errors"
	"strings"
)

const minJWTSecretLength = 32

func isProduction(env string) bool {
	return strings.EqualFold(strings.TrimSpace(env), "production")
}

func validateProductionConfig(lookup func(string) string) error {
	var problems []string

	secret := strings.TrimSpace(lookup("JWT_SECRET"))
	if len(secret) < minJWTSecretLength || secret == "change-me" {
		problems = append(problems, "JWT_SECRET must be at least 32 characters and not the default")
	}

	origins := strings.TrimSpace(lookup("ALLOWED_ORIGINS"))
	if origins == "" || strings.Contains(origins, "*") {
		problems = append(problems, "ALLOWED_ORIGINS must list explicit origins")
	}

	if strings.TrimSpace(lookup("APP_PUBLIC_URL")) == "" {
		problems = append(problems, "APP_PUBLIC_URL is required")
	}

	if len(problems) > 0 {
		return errors.New("invalid production config: " + strings.Join(problems, "; "))
	}
	return nil
}
