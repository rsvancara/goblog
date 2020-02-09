package main

import (
	"blog/blog/config"
	"blog/blog/metrics"
	"blog/blog/routes"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	fmt.Println("== Starting Service ==")

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println("== Initializing Configuration ==")
	fmt.Printf("Database URI: %s\n", cfg.Dburi)
	fmt.Printf("Cache URI: %s\n", cfg.Cacheuri)

	middleware := metrics.NewPrometheusMiddleware(metrics.Opts{})

	r := routes.GetRoutes()

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
			log.Println(err)
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
	log.Println("shutting down")
	os.Exit(0)
}
