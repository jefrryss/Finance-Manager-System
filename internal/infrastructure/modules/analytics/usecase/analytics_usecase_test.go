package usecase

import (
	"testing"
	"time"
)

func TestResolveDatesMonthDefault(t *testing.T) {
	uc := &AnalyticsUseCase{}
	s, e, err := uc.resolveDates(nil, nil, "month")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if s.Day() != 1 {
		t.Fatalf("start day must be 1, got %d", s.Day())
	}
	if e.Before(s) {
		t.Fatalf("end before start")
	}
}

func TestResolveDatesInvalidPeriod(t *testing.T) {
	uc := &AnalyticsUseCase{}
	_, _, err := uc.resolveDates(nil, nil, "year")
	if err != ErrInvalidPeriod {
		t.Fatalf("expected ErrInvalidPeriod, got %v", err)
	}
}

func TestResolveDatesCustomRange(t *testing.T) {
	uc := &AnalyticsUseCase{}
	start := time.Now().Add(-24 * time.Hour).UTC()
	end := time.Now().UTC()
	s, e, err := uc.resolveDates(&start, &end, "")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !s.Equal(start) || !e.Equal(end) {
		t.Fatalf("unexpected returned range")
	}
}

