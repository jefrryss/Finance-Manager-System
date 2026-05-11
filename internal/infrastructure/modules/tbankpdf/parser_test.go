package tbankpdf

import (
	"testing"
	"time"
)

func TestParseMoneyToMinor(t *testing.T) {
	v, err := parseMoneyToMinor("1 234,56 ₽")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if v != 123456 {
		t.Fatalf("unexpected value: %d", v)
	}
}

func TestBuildExternalIDStable(t *testing.T) {
	ts := time.Date(2026, 5, 10, 10, 0, 0, 0, time.UTC)
	id1 := buildExternalID(ts, 5000, false, "Market", "1234")
	id2 := buildExternalID(ts, 5000, false, "Market", "1234")
	if id1 != id2 {
		t.Fatalf("external id must be deterministic")
	}
}

func TestParseStatementInvalidData(t *testing.T) {
	_, err := ParseStatement([]byte("not-a-pdf"))
	if err == nil {
		t.Fatalf("expected error")
	}
}

