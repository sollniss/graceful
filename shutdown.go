package graceful

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	_ server    = &http.Server{}
	_ tlsserver = &http.Server{}
)

type server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

type tlsserver interface {
	ListenAndServeTLS(certFile string, keyFile string) error
	Shutdown(ctx context.Context) error
}

// Start enables customizable graceful shutdown.
//
// ctx.Done is used to signal a shutdown.
// start is expected to be blocking (i.e. server.ListenAndServe).
// success is the error returned by start on successful shutdown (i.e. http.ErrServerClosed).
// shutdown is the function that initiates a shutdown (i.e. server.Shutdown).
// timeout specifies the time available for the graceful shutdown. If the timeout is exceeded a context.DeadlineExceeded error is returned.
//
// Any error returned by either start or shutdown, that is not success will be wrapped and returned.
func Start(ctx context.Context, start func() error, success error, shutdown func(context.Context) error, timeout time.Duration) error {
	shutdownErr := make(chan error, 1)
	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		shutdownErr <- shutdown(ctx)
	}()

	err := start()
	if !errors.Is(err, success) {
		return fmt.Errorf("graceful: error on ListenAndServe: %w", err)
	}

	if err := <-shutdownErr; err != nil {
		return fmt.Errorf("graceful: error on Shutdown: %w", err)
	}

	return nil
}

// ListenAndServe starts server using ListenAndServe and gracefully shuts it down as soon as ctx is Done.
//
// This function is blocking so additional cleanup can be performed after.
func ListenAndServe(ctx context.Context, server server, timeout time.Duration) error {
	return Start(ctx, server.ListenAndServe, http.ErrServerClosed, server.Shutdown, timeout)
}

// ListenAndServeTLS starts server using ListenAndServeTLS and gracefully shuts it down as soon as ctx is Done.
//
// This function is blocking so additional cleanup can be performed after.
func ListenAndServeTLS(ctx context.Context, server tlsserver, certFile string, keyFile string, timeout time.Duration) error {
	start := func() error {
		return server.ListenAndServeTLS(certFile, keyFile)
	}

	return Start(ctx, start, http.ErrServerClosed, server.Shutdown, timeout)
}

// NotifyShutdown is a utility function that calls
//
//	signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
//
// and returns the context only.
func NotifyShutdown() context.Context {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	return ctx
}
