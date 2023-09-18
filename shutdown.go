package graceful

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var (
	_ server = &http.Server{}
)

type server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// HTTPServer starts server and gracefully shuts it down as soon as ctx is Done.
//
// This function is blocking so additional cleanup can be performed after.
func HTTPServer(ctx context.Context, server server, timeout time.Duration) error {
	shutdownErr := make(chan error, 1)
	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		shutdownErr <- server.Shutdown(ctx)
	}()

	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("graceful: error on ListenAndServe: %w", err)
	}

	if err := <-shutdownErr; err != nil {
		return fmt.Errorf("graceful: error on Shutdown: %w", err)
	}

	return nil
}

/*
func HTTPServer2(server *http.Server, timeout time.Duration, sig ...os.Signal) error {
	listenErr := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			listenErr <- fmt.Errorf("graceful: error on ListenAndServe: %s", err)
		}
	}()

	if len(sig) == 0 {
		sig = append(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	}
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, sig...)
	defer close(stopCh)

	<-stopCh

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful: error on Shutdown: %w", err)
	}
	return nil
}

func ShutdownHandler(start func() error, shutdown func(context.Context) error, timeout time.Duration, sig ...os.Signal) error {
	go func() {
		err := start()
		if err != nil {
			fmt.Printf("graceful: error on start: %s", err)
		} else {
			fmt.Printf("graceful: shutdown ok")
		}
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, sig...)
	defer close(stopCh)
	<-stopCh

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := shutdown(ctx)
	if err != nil {
		return fmt.Errorf("graceful: error on shutdown: %w", err)
	}
	return nil
}
*/
