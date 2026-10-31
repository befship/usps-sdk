//go:build ignore

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/befship/usps-sdk/usps"
)

func main() {
	client := usps.New(usps.Config{
		ClientID:     os.Getenv("USPS_CLIENT_ID"),
		ClientSecret: os.Getenv("USPS_CLIENT_SECRET"),
		Environment:  usps.EnvTEM,
	})
	api, err := client.Addresses(context.Background(), usps.Creds{})
	if err != nil {
		panic(err)
	}
	fmt.Println("base", usps.BaseURL(client.Environment()), "api=", api != nil)
}
