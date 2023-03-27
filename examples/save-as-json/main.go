package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jonas-meyer/goopla/goopla"
	"github.com/rs/zerolog/log"
	"os"
)

var ctx = context.Background()

func main() {
	client, err := goopla.NewClient(goopla.Credentials{}, goopla.FromEnv)
	if err != nil {
		log.Fatal().Msgf("%s", err)
	}

	listingOptions := &goopla.ListingOptions{Area: "London", Minimum_beds: 2, Maximum_beds: 2, Order_by: "age", Page_size: 30, Listing_status: "rent"}

	listings, response, err := client.Listing.Get(ctx, listingOptions)
	print(response)

	if err != nil {
		log.Fatal().Msgf("%s", err)
	}

	for _, listing := range listings.Listings {
		listingJson, err := json.MarshalIndent(&listing, "", "  ")
		if err != nil {
			log.Fatal().Msgf("%s", err)
		}
		err = os.WriteFile(fmt.Sprintf("examples/save-as-json/%s.json", listing.ListingID), listingJson, 0644)
		if err != nil {
			log.Fatal().Msgf("%s", err)
		}
	}
}
