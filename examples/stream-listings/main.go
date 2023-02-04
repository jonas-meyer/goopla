package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonas-meyer/goopla/goopla"
	"github.com/rs/zerolog/log"
)

var ctx = context.Background()

func main() {
	sig := make(chan os.Signal, 1)
	defer close(sig)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	client, err := goopla.NewClient(goopla.Credentials{}, goopla.FromEnv)
	if err != nil {
		log.Fatal().Msgf("%s", err)
	}

	listings, errs, stop := client.Stream.Listings(&goopla.ListingOptions{Area: "Oxford", Minimum_beds: 2, Maximum_beds: 2}, goopla.StreamInterval(time.Minute*5))
	defer stop()

	timer := time.NewTimer(time.Minute * 1000)
	defer timer.Stop()

	for {
		select {
		case listing, ok := <-listings:
			if !ok {
				return
			}
			log.Info().Msgf("Received listing: %s", listing.ListingID)
		case err, ok := <-errs:
			if !ok {
				return
			}
			log.Err(err)
		case rcvSig, ok := <-sig:
			if !ok {
				return
			}
			fmt.Printf("Stopping due to %s signal.", rcvSig)
			return
		case <-timer.C:
			return
		}
	}
}
