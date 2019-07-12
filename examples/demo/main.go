package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bitfield/checkly"
)

func main() {
	apiKey := os.Getenv("CHECKLY_API_KEY")
	if apiKey == "" {
		log.Fatal("no CHECKLY_API_KEY set")
	}
	client := checkly.NewClient(apiKey)
	check := checkly.Check{
		Name:      "My Awesome Check",
		Type:      checkly.TypeAPI,
		Frequency: 5,
		Activated: true,
		Locations: []string{"eu-west-1"},
		Request: checkly.Request{
			Method: http.MethodGet,
			URL:    "http://example.com",
		},
	}
	ID, err := client.Create(check)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("New check created with ID %s\n", ID)
}
