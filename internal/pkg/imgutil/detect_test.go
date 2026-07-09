package imgutil

import (
	"testing"
)

var (
	jpegHeader = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	pngHeader  = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	gifHeader  = []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}
	pdfHeader  = []byte{0x25, 0x50, 0x44, 0x46}
	elfHeader  = []byte{0x7F, 0x45, 0x4C, 0x46}
	textHeader = []byte("Hello, world!")
)

func TestValidateImageHeader(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{"JPEG", jpegHeader, false},
		{"PNG", pngHeader, false},
		{"GIF", gifHeader, false},
		{"PDF", pdfHeader, true},
		{"ELF", elfHeader, true},
		{"TEXT", textHeader, true},
		{"empty", []byte{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImageHeader(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateImageHeader() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
