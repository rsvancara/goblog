package main

import (
	"context"
	"flag"
	"fmt"

	//"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rsvancara/goblog/internal/config"
	"github.com/rsvancara/goblog/internal/db"
	"github.com/rsvancara/goblog/internal/handlers"
	"github.com/rsvancara/goblog/internal/metrics"
	"github.com/rsvancara/goblog/internal/routes"

	mediadao "github.com/rsvancara/goblog/internal/dao/media"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	debug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()

	fmt.Println("== Starting Service ==")

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Can not get configuration")
	}

	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Info().Str("service", "main").Msgf("Loading mongo client")
	dbclient, err := db.GetMongoClient(&cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Error getting mongo client")
	}

	// Test DAO
	var mediaDAO mediadao.MediaDAO

	err = mediaDAO.Initialize(dbclient, &cfg)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error initialzing media data access object ")

	}

	log.Info().Str("service", "main").Msgf("Starting up")

	log.Info().Str("service", "main").Msgf("Loading mongo client")

	fmt.Println("== Initializing Configuration ==")
	fmt.Printf("Database URI: %s\n", cfg.Dburi)
	fmt.Printf("Cache URI: %s\n", cfg.Cacheuri)

	log.Info().Str("service", "main").Msgf("Populating configuration and mongo client into context")
	hctx := handlers.CTXHandlerContext(&cfg, dbclient)

	middleware := metrics.NewPrometheusMiddleware(metrics.Opts{})

	r := routes.GetRoutes(hctx)

	r.Handle("/metrics", promhttp.Handler())
	r.Use(middleware.InstrumentHandlerDuration)

	fmt.Println("Now serving requests")

	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:5000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error().Err(err).Msg("Error starting up HTTP Listener")
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.

	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Info().Msg("Shutting down server")
	os.Exit(0)
}
