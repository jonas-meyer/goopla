package goopla

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type StreamService struct {
	client *Client
}

func (s *StreamService) Listings(listingOpts *ListingOptions, streamOpts ...StreamOpt) (<-chan Listing, <-chan error, func()) {
	streamConfig := &streamConfig{
		Interval:       defaultStreamInterval,
		DiscardInitial: false,
	}

	for _, opt := range streamOpts {
		opt(streamConfig)
	}
	listingOpts.Order_by = "age"

	ticker := time.NewTicker(streamConfig.Interval)
	listingsCh := make(chan Listing)
	errsCh := make(chan error)

	var once sync.Once
	stop := func() {
		once.Do(func() {
			ticker.Stop()
			close(listingsCh)
			close(errsCh)
		})
	}

	ids := set{}

	go func() {
		defer stop()

		for ; ; <-ticker.C {
			log.Info().Msgf("Getting newest listings")
			listings, _, err := s.client.Listing.Get(context.Background(), listingOpts)
			if err != nil {
				errsCh <- err
				continue
			}

			for _, listing := range listings.Listings {
				id := listing.ListingID

				if ids.Exists(id) {
					break
				}
				ids.Add(id)

				if streamConfig.DiscardInitial {
					streamConfig.DiscardInitial = false
					break
				}

				listingsCh <- listing
			}
			log.Info().Msgf("No new listings available")
		}
	}()
	return listingsCh, errsCh, stop
}

type set map[string]struct{}

func (s set) Add(i string) {
	s[i] = struct{}{}
}

func (s set) Delete(i string) {
	delete(s, i)
}

func (s set) Len() int {
	return len(s)
}

func (s set) Exists(i string) bool {
	_, ok := s[i]
	return ok
}
