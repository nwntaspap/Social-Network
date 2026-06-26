package categories

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
		{"name is whitelisted", "name", true},
		{"created_by is whitelisted", "created_by", true},
		{"created_at is whitelisted", "created_at", true},
		{"title is not whitelisted", "title", false},
		{"updated_at is not whitelisted", "updated_at", false},
		{"vote_score is not whitelisted", "vote_score", false},
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
