package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/checkly/checkly-go-sdk"
)

var alertSettings = checkly.AlertSettings{
	EscalationType: checkly.RunBased,
	RunBasedEscalation: checkly.RunBasedEscalation{
		FailedRunThreshold: 1,
	},
	TimeBasedEscalation: checkly.TimeBasedEscalation{
		MinutesFailingThreshold: 5,
	},
	Reminders: checkly.Reminders{
		Interval: 5,
	},
}

var dashboard = checkly.Dashboard{
	CustomURL:      "string",
	CustomDomain:   "string",
	Logo:           "string",
	Header:         "string",
	Width:          "FULL",
	RefreshRate:    60,
	PaginationRate: 30,
	Tags:           []string{"string"},
	HideTags:       false,
}

var mw = checkly.MaintenanceWindow{
	Name:           "TEST",
	StartsAt:       "2012-08-24",
	EndsAt:         "2014-08-24",
	RepeatUnit:     "MONTH",
	RepeatEndsAt:   "2013-08-24",
	RepeatInterval: 10,
	CreatedAt:      "2013-08-24",
	UpdatedAt:      "2015-08-24",
	Tags:           []string{"string"},
}

var privateLocation = checkly.PrivateLocation{
	Name:     "Private Location",
	SlugName: "new-private-location",
	Icon:     "location",
}

var apiCheck = checkly.APICheck{
	Name:                 "My API Check",
	Frequency:            10,
	FrequencyOffset:      1,
	DegradedResponseTime: 5000,
	MaxResponseTime:      15000,
	Activated:            true,
	Muted:                false,
	ShouldFail:           false,
	Locations: []string{
		"eu-west-1",
		"ap-northeast-2",
	},
	Tags: []string{
		"foo",
		"bar",
	},
	AlertSettings:          &alertSettings,
	UseGlobalAlertSettings: false,
	Request: checkly.Request{
		Method: http.MethodGet,
		URL:    "http://example.com",
		Headers: []checkly.KeyValue{
			{
				Key:   "X-Test",
				Value: "foo",
			},
		},
		QueryParameters: []checkly.KeyValue{
			{
				Key:   "query",
				Value: "foo",
			},
		},
		Assertions: []checkly.Assertion{
			{
				Source:     checkly.StatusCode,
				Comparison: checkly.Equals,
				Target:     "200",
			},
		},
		Body:     "",
		BodyType: "NONE",
	},
}

var browserCheck = checkly.BrowserCheck{
	Name:           "My Browser Check",
	Frequency:      10,
	Activated:      true,
	Muted:          false,
	ShouldFail:     false,
	SSLCheckDomain: "www.acme.com",
	Locations:      []string{"eu-west-1"},
	AlertSettings:  &alertSettings,
	Script: `const assert = require("chai").assert;
  const puppeteer = require("puppeteer");

  const browser = await puppeteer.launch();
  const page = await browser.newPage();
  await page.goto("https://example.com");
  const title = await page.title();

  assert.equal(title, "Example Site");
  await browser.close();`,
	EnvironmentVariables: []checkly.EnvironmentVariable{
		{
			Key:   "HELLO",
			Value: "Hello world",
		},
	},
}

var group = checkly.Group{
	Name:        "test",
	Activated:   true,
	Muted:       false,
	Tags:        []string{"auto"},
	Locations:   []string{"eu-west-1"},
	Concurrency: 3,
	APICheckDefaults: checkly.APICheckDefaults{
		BaseURL: "example.com/api/test",
		Headers: []checkly.KeyValue{
			{
				Key:   "X-Test",
				Value: "foo",
			},
		},
		QueryParameters: []checkly.KeyValue{
			{
				Key:   "query",
				Value: "foo",
			},
		},
		Assertions: []checkly.Assertion{
			{
				Source:     checkly.StatusCode,
				Comparison: checkly.Equals,
				Target:     "200",
			},
		},
		BasicAuth: checkly.BasicAuth{
			Username: "user",
			Password: "pass",
		},
	},
	EnvironmentVariables: []checkly.EnvironmentVariable{
		{
			Key:   "ENVTEST",
			Value: "Hello world",
		},
	},
	DoubleCheck:            true,
	UseGlobalAlertSettings: false,
	AlertSettings: checkly.AlertSettings{
		EscalationType: checkly.RunBased,
		RunBasedEscalation: checkly.RunBasedEscalation{
			FailedRunThreshold: 1,
		},
		TimeBasedEscalation: checkly.TimeBasedEscalation{
			MinutesFailingThreshold: 5,
		},
		Reminders: checkly.Reminders{
			Amount:   0,
			Interval: 5,
		},
	},
	LocalSetupScript:    "setup-test",
	LocalTearDownScript: "teardown-test",
}

var heartbeatCheck = checkly.HeartbeatMonitor{
	Name:          "My Heartbeat Check",
	Activated:     true,
	Muted:         false,
	AlertSettings: alertSettings,
	Heartbeat: checkly.Heartbeat{
		Period:     30,
		PeriodUnit: "seconds",
		Grace:      0,
		GraceUnit:  "seconds",
	},
}

func main() {
	// An unset CHECKLY_API_URL leaves the production API. WithAuthFromEnv
	// requires both CHECKLY_API_KEY and CHECKLY_ACCOUNT_ID.
	client, err := checkly.NewClient(
		checkly.WithBaseURL(os.Getenv("CHECKLY_API_URL")),
		checkly.WithAuthFromEnv(),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	group, err := client.CreateGroup(ctx, group)
	if err != nil {
		log.Fatalf("creating group: %v", err)
	}
	fmt.Printf("New check group created with ID %d\n", group.ID)

	gotAPICheck, err := client.CreateAPICheck(ctx, apiCheck)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("New API check created with ID %s\n", gotAPICheck.ID)

	gotBrowserCheck, err := client.CreateBrowserCheck(ctx, browserCheck)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("New browser check created with ID %s\n", gotBrowserCheck.ID)

	// HB
	gotHbCheck, err := client.CreateHeartbeatMonitor(ctx, heartbeatCheck)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("New check created with ID %s\n", gotHbCheck.ID)
}
