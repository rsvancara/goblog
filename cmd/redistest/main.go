package main

import (
	"flag"
	"fmt"
	"goblog/internal/cache"
	"goblog/internal/config"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

	var cache cache.Cache

	err = cache.InitPool(cfg)
	if err != nil {
		log.Error().Err(err).Msg("Error initializing pool")
	}

	cache.Ping()

	keys, err := cache.GetAllVals("*")
	if err != nil {
		fmt.Println(err)
	}

	var wg sync.WaitGroup

	for _, v := range keys {

		wg.Add(1)

		go func(value string) {
			defer wg.Done()

			fmt.Println(value)
			resp, err := cache.GetKey(value)
			if err != nil {
				log.Error().Err(err).Msg("error finding key")
			}

			fmt.Println(resp)
			stats, _ := cache.GetPoolStatus()
			fmt.Println(stats)
		}(v)

		wg.Wait()
	}

}
