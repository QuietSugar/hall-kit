package dockerage

import (
	"errors"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	utc := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		in   string
		want time.Time
	}{
		{"utc", "2026-10-17T12:00:00Z", utc},
		{"positive offset", "2026-10-17T20:00:00+08:00", utc},
		{"negative offset", "2026-10-17T07:00:00-05:00", utc},
		{"nanoseconds", "2026-10-17T20:00:00.744706645+08:00", utc.Add(744706645 * time.Nanosecond)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.in)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.in, err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("Parse(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseInvalid(t *testing.T) {
	for _, in := range []string{"", "not-a-time", "2026-13-01T00:00:00Z", "2026-10-17"} {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) returned nil error, want error", in)
		}
	}
}

func TestSecondsSince(t *testing.T) {
	now := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		in   string
		want int64
	}{
		{"same instant", "2026-10-17T12:00:00Z", 0},
		{"one second ago", "2026-10-17T11:59:59Z", 1},
		{"ninety seconds ago", "2026-10-17T11:58:30Z", 90},
		{"sub-second floors", "2026-10-17T11:59:59.999999999Z", 0},
		{"positive offset", "2026-10-17T19:59:00+08:00", 60},
		{"one day ago", "2026-10-16T12:00:00Z", 86400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, err := Parse(tt.in)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.in, err)
			}
			got, err := SecondsSince(ts, now)
			if err != nil {
				t.Fatalf("SecondsSince(%v, %v) returned error: %v", ts, now, err)
			}
			if got != tt.want {
				t.Errorf("SecondsSince(%q, now) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestSecondsSinceFuture(t *testing.T) {
	now := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	ts, err := Parse("2026-10-17T12:00:01Z")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if _, err := SecondsSince(ts, now); !errors.Is(err, ErrFuture) {
		t.Fatalf("SecondsSince(future) error = %v, want ErrFuture", err)
	}
}
