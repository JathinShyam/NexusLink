package validation

import "testing"

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "https ok", input: "https://example.com/path", want: "https://example.com/path"},
		{name: "http ok", input: "http://example.com", want: "http://example.com"},
		{name: "trims space", input: "  https://example.com  ", want: "https://example.com"},
		{name: "empty", input: "", wantErr: true},
		{name: "ftp", input: "ftp://example.com", wantErr: true},
		{name: "no host", input: "https://", wantErr: true},
		{name: "localhost", input: "http://localhost/admin", wantErr: true},
		{name: "private ip", input: "http://192.168.1.1", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateURL(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateShortCode(t *testing.T) {
	if err := ValidateShortCode("abc"); err != nil {
		t.Errorf("abc should be valid: %v", err)
	}
	if err := ValidateShortCode("ab"); err == nil {
		t.Error("expected error for too-short code")
	}
	if err := ValidateShortCode("bad alias"); err == nil {
		t.Error("expected error for spaces")
	}
}
