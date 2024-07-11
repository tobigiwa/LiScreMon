package main

import (
	webserver "agent"
	helperFuncs "pkg/helper"

	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {

	// logging
	logger, logFile, err := helperFuncs.Logger("webserver.log")
	if err != nil {
		log.Fatalln(err) // exit
	}
	defer logFile.Close()

	slog.SetDefault(logger)

	app, err := webserver.NewApp(logger)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			log.Fatalln("daemon service is not running", err)
		}
		log.Fatalln("error creating app:", err)
	}

	_, err = app.CheckDaemonService()
	if err != nil {
		log.Fatalln("error connecting to daemon service:", err)
	}

	server := &http.Server{
		Addr:     ":8080",
		Handler:  app.Routes(),
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Println("Server is running on port http://127.0.0.1:8080/screentime")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Println("Server error:", err)
		}
	}()

	<-done
	close(done)

	if err := app.CloseDaemonConnection(); err != nil {
		fmt.Println("error closing socket connection with daemon, error:", err)
	}

	if err := server.Shutdown(context.TODO()); err != nil {
		fmt.Printf("Graceful server shutdown Failed:%+v\n", err)
	}

	fmt.Println("SERVER STOPPED GRACEFULLY")
}
