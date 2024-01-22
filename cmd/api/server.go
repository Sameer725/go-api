package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) server() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	shutDownError := make(chan error)

	// start a background goroutine
	go func() {
		//create a quit channel which carries os.Signal
		quit := make(chan os.Signal, 1)

		//signal.Notify to listen incoming sigint and sigterm and relay them to quit channel.
		//any other signal will not be caught and will retain their default behavior
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		// read signal from quit channel, this code will block until a signal is received
		s := <-quit

		app.logger.Info("caught signal", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		//returns nil if graceful shutdown is success
		// error if problem closing listeners
		shutDownError <- srv.Shutdown(ctx)

		// exit tha app with success status code
		os.Exit(0)
	}()

	app.logger.Info("Server Starting", "addr", srv.Addr, "env", app.config.env)

	//calling ShutDown will call ListenAndServe() to immediately return a http.ErrServerClosed indicating graceful shutdown started
	err := srv.ListenAndServe()

	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutDownError

	if err != nil {
		return err
	}

	//at this point we know graceful shutdown is successful
	app.logger.Info("stopped server", "addr", srv.Addr)

	return nil
}
