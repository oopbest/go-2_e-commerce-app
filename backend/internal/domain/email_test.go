package domain

import "testing"

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "already normalized",
			input: "admin@example.com",
			want:  "admin@example.com",
		},
		{
			name:  "converts uppercase to lowercase",
			input: "Admin@Example.COM",
			want:  "admin@example.com",
		},
		{
			name:  "removes surrounding spaces",
			input: "  admin@example.com  ",
			want:  "admin@example.com",
		},
		{
			name:  "normalizes case and whitespace",
			input: "  Admin@Example.COM  ",
			want:  "admin@example.com",
		},
		{
			name:  "removes tabs and newlines",
			input: "\tAdmin@Example.COM\n",
			want:  "admin@example.com",
		},
		{
			name:  "empty input remains empty",
			input: "",
			want:  "",
		},
		{
			name:  "whitespace-only input becomes empty",
			input: "   ",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeEmail(tt.input)

			if got != tt.want {
				t.Errorf(
					"NormalizeEmail(%q) = %q, want %q",
					tt.input,
					got,
					tt.want,
				)
			}
		})
	}
}
