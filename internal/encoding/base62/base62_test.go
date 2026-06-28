package base62

import (
	"errors"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	cases := []uint64{0, 1, 61, 62, 63, 3844, 1_000_000, 3_521_614_606_208}
	for _, n := range cases {
		encoded := Encode(n)
		decoded, err := Decode(encoded)
		if err != nil {
			t.Fatalf("Decode(%q) error = %v", encoded, err)
		}
		if decoded != n {
			t.Errorf("round trip %d: got %d from %q", n, decoded, encoded)
		}
	}
}

func TestEncodeKnownValues(t *testing.T) {
	if got := Encode(0); got != "0" {
		t.Errorf("Encode(0) = %q, want 0", got)
	}
	if got := Encode(61); got != "Z" {
		t.Errorf("Encode(61) = %q, want Z", got)
	}
	if got := Encode(62); got != "10" {
		t.Errorf("Encode(62) = %q, want 10", got)
	}
}

func TestDecodeInvalid(t *testing.T) {
	_, err := Decode("")
	if err != ErrEmptyInput {
		t.Errorf("Decode(\"\") error = %v, want ErrEmptyInput", err)
	}

	_, err = Decode("abc!")
	if !errors.Is(err, ErrInvalidChar) {
		t.Errorf("Decode(\"abc!\") error = %v, want ErrInvalidChar", err)
	}
}
