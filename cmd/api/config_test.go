package main

import (
	"strings"
	"testing"
)

func TestValidateProductionConfig(t *testing.T) {
	valid := map[string]string{
		"JWT_SECRET":      strings.Repeat("s", 32),
		"ALLOWED_ORIGINS": "https://oportunidades.social,https://admin.oportunidades.social",
		"APP_PUBLIC_URL":  "https://oportunidades.social",
	}

	cases := map[string]struct {
		override map[string]string
		wantErr  string
	}{
		"valid":              {},
		"missing secret":     {override: map[string]string{"JWT_SECRET": ""}, wantErr: "JWT_SECRET"},
		"short secret":       {override: map[string]string{"JWT_SECRET": "short"}, wantErr: "JWT_SECRET"},
		"default secret":     {override: map[string]string{"JWT_SECRET": "change-me"}, wantErr: "JWT_SECRET"},
		"missing origins":    {override: map[string]string{"ALLOWED_ORIGINS": ""}, wantErr: "ALLOWED_ORIGINS"},
		"wildcard origins":   {override: map[string]string{"ALLOWED_ORIGINS": "*"}, wantErr: "ALLOWED_ORIGINS"},
		"missing public url": {override: map[string]string{"APP_PUBLIC_URL": ""}, wantErr: "APP_PUBLIC_URL"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			env := map[string]string{}
			for k, v := range valid {
				env[k] = v
			}
			for k, v := range tc.override {
				env[k] = v
			}

			err := validateProductionConfig(func(key string) string { return env[key] })
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error mentioning %s, got %v", tc.wantErr, err)
			}
		})
	}
}
