package checkly_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/checkly/checkly-go-sdk"
)

// NewClient with WithAuthFromEnv reads the API key and account ID from the
// CHECKLY_API_KEY and CHECKLY_ACCOUNT_ID environment variables, the same ones
// the Checkly CLI and the Terraform provider use.
func ExampleNewClient() {
	client, err := checkly.NewClient(checkly.WithAuthFromEnv())
	if errors.Is(err, checkly.ErrMissingAPIKey) || errors.Is(err, checkly.ErrMissingAccountID) {
		log.Fatal("set CHECKLY_API_KEY and CHECKLY_ACCOUNT_ID: ", err)
	}
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// The static IP ranges checks run from, counted per region.
	ips, err := client.GetStaticIPs(ctx)
	if err != nil {
		log.Fatal(err)
	}
	perRegion := make(map[string]int)
	for _, ip := range ips {
		perRegion[ip.Region]++
	}
	fmt.Printf("%d IP ranges in %d regions\n", len(ips), len(perRegion))
}

// WithAuth takes the credentials as arguments, for example from a secret
// store, and WithHTTPClient sets the HTTP client, here to bound how long a
// single request may take.
func ExampleNewClient_withAuth() {
	client, err := checkly.NewClient(
		checkly.WithAuth(os.Getenv("CHECKLY_API_KEY"), os.Getenv("CHECKLY_ACCOUNT_ID")),
		checkly.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// The addresses checks run from, for allowlisting them in a firewall.
	ips, err := client.GetStaticIPs(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, ip := range ips {
		if ip.Region == "eu-west-1" {
			fmt.Println(ip.Address)
		}
	}
}

// An API check sends an HTTP request every Frequency minutes from each of
// its locations and fails when an assertion does not hold.
func ExampleClient_CreateAPICheck() {
	client, err := checkly.NewClient(checkly.WithAuthFromEnv())
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	check, err := client.CreateAPICheck(ctx, checkly.APICheck{
		Name:      "Example API",
		Frequency: 10,
		Activated: true,
		Locations: []string{"eu-west-1", "us-east-1"},
		Tags:      []string{"example"},
		// The response time limits in milliseconds. Both are always sent,
		// so set them; these are the API's defaults.
		DegradedResponseTime:   5000,
		MaxResponseTime:        20000,
		UseGlobalAlertSettings: true,
		Request: checkly.Request{
			Method:          http.MethodGet,
			URL:             "https://example.com/api/health",
			FollowRedirects: true,
			Headers: []checkly.KeyValue{
				{Key: "Accept", Value: "application/json"},
			},
			Assertions: []checkly.Assertion{
				{Source: checkly.AssertionSourceStatusCode, Comparison: checkly.AssertionComparisonEquals, Target: "200"},
				{Source: checkly.AssertionSourceResponseTime, Comparison: checkly.AssertionComparisonLessThan, Target: "2000"},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	// Delete the check so that the example leaves nothing behind; a real
	// program would keep it.
	defer func() {
		if err := client.DeleteAPICheck(ctx, check.ID); err != nil {
			log.Printf("deleting API check %s: %v", check.ID, err)
		}
	}()

	fmt.Printf("created API check %q with ID %s\n", check.Name, check.ID)
}

// A URL monitor requests a URL with GET and checks the response status,
// which is cheaper than an API check when that is all you need.
func ExampleClient_CreateURLMonitor() {
	client, err := checkly.NewClient(checkly.WithAuthFromEnv())
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	monitor, err := client.CreateURLMonitor(ctx, checkly.URLMonitor{
		Name:                   "Example homepage",
		Frequency:              5,
		Activated:              true,
		Locations:              []string{"eu-west-1", "us-east-1"},
		Tags:                   []string{"example"},
		DegradedResponseTime:   3000,
		MaxResponseTime:        10000,
		UseGlobalAlertSettings: true,
		Request: checkly.URLRequest{
			URL:             "https://example.com/",
			FollowRedirects: true,
			Assertions: []checkly.Assertion{
				{Source: checkly.AssertionSourceStatusCode, Comparison: checkly.AssertionComparisonEquals, Target: "200"},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	// Delete the monitor so that the example leaves nothing behind; a real
	// program would keep it.
	defer func() {
		if err := client.DeleteURLMonitor(ctx, monitor.ID); err != nil {
			log.Printf("deleting URL monitor %s: %v", monitor.ID, err)
		}
	}()

	fmt.Printf("created URL monitor %q with ID %s\n", monitor.Name, monitor.ID)
}

// GetCheck reads a check whose type is not known. A type switch on the result
// reaches the fields of its type.
func ExampleClient_GetCheck() {
	client, err := checkly.NewClient(checkly.WithAuthFromEnv())
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	check, err := client.GetCheck(ctx, "00000000-0000-4000-8000-000000000000")
	if errors.Is(err, checkly.ErrUnsupportedCheckType) {
		fmt.Println("the SDK has no type for this check:", err)
		return
	}
	if err != nil {
		log.Fatal(err)
	}

	switch c := check.(type) {
	case *checkly.APICheck:
		fmt.Printf("API check %q requests %s\n", c.Name, c.Request.URL)
	case *checkly.HeartbeatMonitor:
		fmt.Printf("heartbeat monitor %q expects a ping every %d %s\n", c.Name, c.Heartbeat.Period, c.Heartbeat.PeriodUnit)
	default:
		fmt.Printf("another check type: %T\n", check)
	}
}

// Match an error with errors.Is and the sentinel errors to branch on the
// kind of failure, and use errors.As to read the details of an *APIError.
func ExampleAPIError() {
	client, err := checkly.NewClient(checkly.WithAuthFromEnv())
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = client.GetAPICheck(ctx, "00000000-0000-4000-8000-000000000000")
	switch {
	case err == nil:
		fmt.Println("the check exists")
	case errors.Is(err, checkly.ErrNotFound):
		fmt.Println("the check does not exist")
	case errors.Is(err, checkly.ErrValidation):
		var apiErr *checkly.APIError
		if errors.As(err, &apiErr) {
			fmt.Println("invalid request fields:", apiErr.Details)
		}
	case errors.Is(err, checkly.ErrRateLimited):
		fmt.Println("rate limited; try again later")
	default:
		var apiErr *checkly.APIError
		if errors.As(err, &apiErr) {
			fmt.Printf("the API answered %d %s: %s\n", apiErr.StatusCode, apiErr.Code, apiErr.Message)
		} else {
			// No response, for example a network failure or an expired
			// context.
			fmt.Println("request failed:", err)
		}
	}
}
