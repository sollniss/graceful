package graceful

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

type mockServer struct {
	errListen    error
	errShutdown  error
	stopping     bool
	stopped      bool
	shutdownTime time.Duration
}

func (s *mockServer) ListenAndServe() error {
	if s.errListen != nil {
		return s.errListen
	}
	for !s.stopping {
		time.Sleep(10 * time.Millisecond)
	}
	s.stopped = true
	// http.Server does never return nil
	return http.ErrServerClosed
}

func (s *mockServer) Shutdown(ctx context.Context) error {
	s.stopping = true
	for !s.stopped {
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(s.shutdownTime)
	if s.errShutdown != nil {
		return s.errShutdown
	}
	return ctx.Err()
}

func TestListenAndServeError(t *testing.T) {
	errExp := errors.New("error")
	s := &mockServer{errListen: errExp, errShutdown: nil}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := ListenAndServe(ctx, s, 50*time.Millisecond)
	if !errors.Is(err, errExp) {
		t.Fatalf("invalid error, got: %s, expected: %s", err, errExp)
	}
}

func TestShutdownError(t *testing.T) {
	errExp := errors.New("error")
	s := &mockServer{errListen: nil, errShutdown: errExp}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := ListenAndServe(ctx, s, 50*time.Millisecond)
	if !errors.Is(err, errExp) {
		t.Fatalf("invalid error, got: %s, expected: %s", err, errExp)
	}
}

func TestShutdownTimeout(t *testing.T) {
	s := &mockServer{errListen: nil, errShutdown: nil, shutdownTime: 1 * time.Second}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := ListenAndServe(ctx, s, 50*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("invalid error, got: %s, expected: %s", err, context.DeadlineExceeded)
	}
}
