package shortener

import (
	"context"
	"testing"
)

func TestBase62Shortener_Generate(t *testing.T) {
	s := NewBase62()
	ctx := context.Background()

	code, err := s.Generate(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code == "" {
		t.Fatal("code should not be empty")
	}
	t.Logf("code: %s", code)
}

func TestBase62Shortener_Deterministic(t *testing.T) {
	s := NewBase62()
	ctx := context.Background()

	url := "https://example.com/very/long/url/with/query?param=value&other=123"
	code1, _ := s.Generate(ctx, url)
	code2, _ := s.Generate(ctx, url)

	if code1 != code2 {
		t.Fatalf("expected deterministic code: got %q and %q", code1, code2)
	}
}

func TestBase62Shortener_DifferentURLs_DifferentCodes(t *testing.T) {
	s := NewBase62()
	ctx := context.Background()

	code1, _ := s.Generate(ctx, "https://example.com/a")
	code2, _ := s.Generate(ctx, "https://example.com/b")

	if code1 == code2 {
		t.Fatal("different URLs should produce different codes (with very high probability)")
	}
}

func TestEncode(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{61, "Z"},
		{62, "10"},
		{63, "11"},
		{3844, "100"}, // 62^2
	}

	for _, tc := range tests {
		result := encode(tc.input)
		if result != tc.expected {
			t.Errorf("encode(%d) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}