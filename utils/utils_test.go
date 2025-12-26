package utils

import (
	"errors"
	"testing"
	"time"
)

func TestRetry_AlwaysFail(t *testing.T) {
	attempt := 0

	_, err := Retry[int](3, 50*time.Millisecond, func() (int, error) {
		attempt++
		return 0, errors.New("always fail")
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if attempt != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempt)
	}
}

func TestRetry_SuccessAfterRetries(t *testing.T) {
	attempt := 0

	result, err := Retry[int](5, 1000*time.Millisecond, func() (int, error) {
		attempt++
		if attempt < 5 {
			return 0, errors.New("fail")
		}
		return 42, nil
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if result != 42 {
		t.Fatalf("expected 42, got %d", result)
	}

	if attempt != 5 {
		t.Fatalf("expected 3 attempts, got %d", attempt)
	}
}
