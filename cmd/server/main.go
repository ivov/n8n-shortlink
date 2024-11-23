package main

import (
	"context"
	"errors"
	"fmt"
	nativeLog "log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/ivov/n8n-shortlink/internal/api"
	"github.com/ivov/n8n-shortlink/internal/config"
	"github.com/ivov/n8n-shortlink/internal/db"
	"github.com/ivov/n8n-shortlink/internal/log"
	"github.com/ivov/n8n-shortlink/internal/services"
)

var (
	// SHA and buildTime are set at compile time
	commitSha string
	buildTime string
)

func main() {
	cfg := config.NewConfig(commitSha)

	if *cfg.MetadataMode {
		fmt.Println(commitSha)
		fmt.Printf("Commit SHA:\t%s\n", commitSha)
		fmt.Printf("Build time:\t%s\n", buildTime)
		os.Exit(0)
	}

	config.SetupDotDir()

	// ------------
	//    logger
	// ------------

	logger, err := log.NewLogger(cfg.Env)
	if err != nil {
		nativeLog.Fatalf("failed to init logger: %v", err)
	}

	logger.ReportEnvs()

	// ------------
	//    sentry
	// ------------

	if cfg.Env == "production" {
		err = sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.Sentry.DSN,
			AttachStacktrace: true,
		})
		if err != nil {
			logger.Fatal(err)
			os.Exit(1)
		}

		defer sentry.Flush(2 * time.Second)
	} else {
		logger.Info("sentry disabled in non-production environment")
	}

	// ------------
	//     DB
	// ------------

	db, err := db.Setup(cfg.Env)
	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	}

	defer db.Close()

	// ------------
	//    setup
	// ------------

	api := &api.API{
		Config:           &cfg,
		Logger:           &logger,
		ShortlinkService: &services.ShortlinkService{DB: db, Logger: &logger},
		VisitService:     &services.VisitService{DB: db, Logger: &logger},
	}

	api.InitMetrics(commitSha)

	server := &http.Server{
		Addr:         cfg.Host + ":" + strconv.Itoa(cfg.Port),
		Handler:      api.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// ------------
	//   shutdown
	// ------------

	// On SIGINT or SIGTERM, give a 5-second grace period for any
	// background tasks to complete before the server shuts down.

	shutdownErrorCh := make(chan error)

	go func() {
		signalCh := make(chan os.Signal, 1)

		signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

		s := <-signalCh

		api.Logger.Info("caught signal", log.Str("signal", s.String()))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := server.Shutdown(ctx)
		if err != nil {
			shutdownErrorCh <- err
		}

		api.Logger.Info("waiting for bkg tasks to complete")

		api.WaitGroup.Wait()
		shutdownErrorCh <- nil
	}()

	// ------------
	//    start
	// ------------

	api.Logger.Info("starting server", log.Str("addr", server.Addr))

	err = server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		api.Logger.Fatal(err)
		os.Exit(1)
	}

	err = <-shutdownErrorCh
	if err != nil {
		api.Logger.Fatal(err)
		os.Exit(1)
	}

	api.Logger.Info("stopped server", log.Str("addr", server.Addr))
}
