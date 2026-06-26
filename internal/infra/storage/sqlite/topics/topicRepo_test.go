package topics

import "testing"

func TestSanitizeOrder(t *testing.T) {
	tests := []struct {
		name  string
		order string
		want  string
	}{
		{"ASC uppercase", "ASC", "ASC"},
		{"DESC uppercase", "DESC", "DESC"},
		{"lowercase asc", "asc", "ASC"},
		{"lowercase desc", "desc", "DESC"},
		{"mixed case", "Asc", "ASC"},
		{"arbitrary value defaults to DESC", "invalid", "DESC"},
		{"empty string defaults to DESC", "", "DESC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeOrder(tt.order)
			if got != tt.want {
				t.Errorf("sanitizeOrder(%q) = %q, want %q", tt.order, got, tt.want)
			}
		})
	}
}

func TestIsOrderByWhitelisted(t *testing.T) {
	tests := []struct {
		name    string
		orderBy string
		want    bool
	}{
		{"created_at is whitelisted", "created_at", true},
		{"updated_at is whitelisted", "updated_at", true},
		{"title is whitelisted", "title", true},
		{"vote_score is whitelisted", "vote_score", true},
		{"name is not whitelisted", "name", false},
		{"created_by is not whitelisted", "created_by", false},
		{"empty string not whitelisted", "", false},
		{"random string not whitelisted", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isOrderByWhitelisted(tt.orderBy)
			if got != tt.want {
				t.Errorf("isOrderByWhitelisted(%q) = %v, want %v", tt.orderBy, got, tt.want)
			}
		})
	}
}
