package main

import "testing"

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "valid email",
			email:   "admin@example.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
		},
		{
			name:    "missing at sign",
			email:   "admin.example.com",
			wantErr: true,
		},
		{
			name:    "missing domain",
			email:   "admin@",
			wantErr: true,
		},
		{
			name:    "display name is not allowed",
			email:   "Admin <admin@example.com>",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmail(tt.email)

			if (err != nil) != tt.wantErr {
				t.Errorf(
					"validateEmail(%q) error = %v, wantErr = %v",
					tt.email,
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestValidatePasswords(t *testing.T) {
	tests := []struct {
		name         string
		password     string
		confirmation string
		wantErr      bool
	}{
		{
			name:         "valid password",
			password:     "securepass12",
			confirmation: "securepass12",
			wantErr:      false,
		},
		{
			name:         "empty password",
			password:     "",
			confirmation: "",
			wantErr:      true,
		},
		{
			name:         "password too short",
			password:     "short",
			confirmation: "short",
			wantErr:      true,
		},
		{
			name:         "passwords do not match",
			password:     "securepass12",
			confirmation: "securepass13",
			wantErr:      true,
		},
		{
			name:         "exactly twelve characters",
			password:     "123456789012",
			confirmation: "123456789012",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePasswords(tt.password, tt.confirmation)

			if (err != nil) != tt.wantErr {
				t.Errorf(
					"validatePasswords() error = %v, wantErr = %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}
