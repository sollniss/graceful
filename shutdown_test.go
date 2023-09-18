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

// TODO: find a way to bind to random port but keep using ListenAndServe
// https://stackoverflow.com/a/43425461
// Mayle loop ports? -> check stdlib what they do if port is :0

// TODO: get rid of mockServer

func TestShutdown(t *testing.T) {
	h := http.NewServeMux()
	h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
	})

	s := &http.Server{
		Handler: h,
		Addr:    "0.0.0.0:8080",
	}

	// Initiate a shutdown 100ms from now.
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Send a request to the server before the shutdown fires.
	go func() {
		// Delay the request for 50ms since the server isn't listening yet.
		time.Sleep(50 * time.Millisecond)
		http.DefaultClient.Get("http://localhost:8080")
	}()

	// The request sent above will process for 1 second.
	// So a shutdown timeout of 200ms triggers a context.DeadlineExceeded error.
	err := HTTPServer(ctx, s, 200*time.Millisecond)
	if errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestListenAndServeError(t *testing.T) {
	errExp := errors.New("error")
	s := &mockServer{errListen: errExp, errShutdown: nil}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := HTTPServer(ctx, s, 50*time.Millisecond)
	if !errors.Is(err, errExp) {
		t.Fatalf("invalid error, got: %s, expected: %s", err, errExp)
	}
}

func TestShutdownError(t *testing.T) {
	errExp := errors.New("error")
	s := &mockServer{errListen: nil, errShutdown: errExp}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := HTTPServer(ctx, s, 50*time.Millisecond)
	if !errors.Is(err, errExp) {
		t.Fatalf("invalid error, got: %s, expected: %s", err, errExp)
	}
}

func TestShutdownTimeout(t *testing.T) {
	s := &mockServer{errListen: nil, errShutdown: nil, shutdownTime: 1 * time.Second}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := HTTPServer(ctx, s, 50*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("invalid error, got: %s, expected: %s", err, context.DeadlineExceeded)
	}
}
