package validator

import "testing"

func TestValidOrder(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantOk  bool
		wantMsg string
	}{
		{"ASC", "ASC", true, ""},
		{"DESC", "DESC", true, ""},
		{"lowercase asc", "asc", true, ""},
		{"lowercase desc", "desc", true, ""},
		{"mixed case", "Asc", true, ""},
		{"invalid value", "invalid", false, "must be a valid order field"},
		{"empty string", "", true, ""},
		{"non-string type", 123, false, InvalidType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, msg := validOrder(tt.value)
			if ok != tt.wantOk || msg != tt.wantMsg {
				t.Errorf("validOrder(%v) = (%v, %q), want (%v, %q)", tt.value, ok, msg, tt.wantOk, tt.wantMsg)
			}
		})
	}
}

func TestValidTopicOrderBy(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantOk  bool
		wantMsg string
	}{
		{"created_at", "created_at", true, ""},
		{"updated_at", "updated_at", true, ""},
		{"title", "title", true, ""},
		{"vote_score", "vote_score", true, ""},
		{"invalid field", "name", false, "must be a valid order by field"},
		{"empty string", "", true, ""},
		{"non-string type", 123, false, InvalidType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, msg := validTopicOrderBy(tt.value)
			if ok != tt.wantOk || msg != tt.wantMsg {
				t.Errorf("validTopicOrderBy(%v) = (%v, %q), want (%v, %q)", tt.value, ok, msg, tt.wantOk, tt.wantMsg)
			}
		})
	}
}

func TestValidCategoryOrderBy(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantOk  bool
		wantMsg string
	}{
		{"name", "name", true, ""},
		{"created_by", "created_by", true, ""},
		{"created_at", "created_at", true, ""},
		{"invalid field", "title", false, "must be a valid order by field"},
		{"empty string", "", true, ""},
		{"non-string type", 123, false, InvalidType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, msg := validCategoryOrderBy(tt.value)
			if ok != tt.wantOk || msg != tt.wantMsg {
				t.Errorf("validCategoryOrderBy(%v) = (%v, %q), want (%v, %q)", tt.value, ok, msg, tt.wantOk, tt.wantMsg)
			}
		})
	}
}
