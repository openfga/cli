package requests_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openfga/cli/internal/requests"
)

func TestRampUpAPIRequests_Success(t *testing.T) {
	t.Parallel()

	var (
		callCount int32
		mutex     sync.Mutex
	)

	requestsList := make([]func() error, 5)
	for i := range requestsList {
		requestsList[i] = func() error {
			mutex.Lock()
			atomic.AddInt32(&callCount, 1)
			mutex.Unlock()

			return nil
		}
	}

	err := requests.RampUpAPIRequests(context.Background(), 10, 20, 2, time.Second, 5, requestsList)
	if err != nil {
		t.Fatalf("expected no error, got %v, %v", err, callCount)
	}

	if callCount != int32(len(requestsList)) { //nolint:gosec
		t.Fatalf("expected %d calls, got %d", len(requestsList), callCount)
	}
}

func TestRampUpAPIRequests_RampUpRate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var (
		callCount int32
		mutex     sync.Mutex
	)

	requestsList := make([]func() error, 10)
	for i := range requestsList {
		requestsList[i] = func() error {
			mutex.Lock()
			atomic.AddInt32(&callCount, 1)
			mutex.Unlock()

			return nil
		}
	}

	startTime := time.Now()

	err := requests.RampUpAPIRequests(ctx, 1, 10, 3, time.Second, 5, requestsList)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	duration := time.Since(startTime).Seconds()
	expectedDuration := float64(len(requestsList)) / 10.0 // maxRPS is 10

	if duration < expectedDuration {
		t.Fatalf("expected duration at least %f seconds, got %f seconds", expectedDuration, duration)
	}
}

func TestRampUpAPIRequests_ContextCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var (
		callCount int32
		mutex     sync.Mutex
	)

	requestsList := make([]func() error, 100)
	for i := range requestsList {
		requestsList[i] = func() error {
			mutex.Lock()
			atomic.AddInt32(&callCount, 1)
			mutex.Unlock()
			time.Sleep(100 * time.Millisecond)

			return nil
		}
	}

	err := requests.RampUpAPIRequests(ctx, 1, 10, 5, time.Second, 5, requestsList)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if callCount == int32(len(requestsList)) { //nolint:gosec
		t.Fatalf("expected fewer than %d calls, got %d", len(requestsList), callCount)
	}
}

func TestRampUpAPIRequests_RequestError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var (
		callCount int32
		mutex     sync.Mutex
	)

	requestsList := make([]func() error, 5)
	for i := range requestsList {
		requestsList[i] = func() error {
			mutex.Lock()
			atomic.AddInt32(&callCount, 1)

			if callCount == 2 {
				mutex.Unlock()

				return errors.New("request error")
			}
			mutex.Unlock()

			return nil
		}
	}

	err := requests.RampUpAPIRequests(ctx, 1, 10, 5, time.Second, 5, requestsList)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if callCount != int32(len(requestsList)) { //nolint:gosec
		t.Fatalf("expected %d calls, got %d", len(requestsList), callCount)
	}
}
