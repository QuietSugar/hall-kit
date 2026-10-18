// Package dockerage computes how old a Docker image is from its creation
// timestamp.
//
// Docker reports an image's creation time as an RFC3339 timestamp, often with
// nanosecond precision and a timezone offset, for example:
//
//	2025-11-07T11:12:09.744706645+08:00
//
// Parse accepts such a value and SecondsSince turns it into the whole number
// of seconds elapsed since a reference time.
package dockerage

import (
	"errors"
	"fmt"
	"time"
)

// ErrFuture reports that the given timestamp is later than the reference time.
var ErrFuture = errors.New("timestamp is in the future")

// Parse parses an RFC3339 timestamp. Fractional seconds are optional.
func Parse(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp %q: %w", s, err)
	}
	return t, nil
}

// SecondsSince returns the whole number of seconds elapsed from t to now,
// rounded down. It returns ErrFuture when t is after now.
func SecondsSince(t, now time.Time) (int64, error) {
	if t.After(now) {
		return 0, ErrFuture
	}
	return int64(now.Sub(t).Seconds()), nil
}
