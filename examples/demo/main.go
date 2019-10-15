package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bitfield/checkly"
)

var alertChannels = checkly.AlertChannels{
	Email: []checkly.AlertEmail{
		{
			Address: "info@example.com",
		},
	},
	Webhook: []checkly.AlertWebhook{
		{
			Name: "test webhook",
			URL:  "http://example.com/webhook",
		},
	},
	Slack: []checkly.AlertSlack{
		{
			URL: "http://slack.com/example",
		},
	},
	SMS: []checkly.AlertSMS{
		{
			Number: "555-5555",
			Name:   "test SMS",
		},
	},
}

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
	SSLCertificates: checkly.SSLCertificates{
		Enabled:        false,
		AlertThreshold: 3,
	},
}

var apiCheck = checkly.Check{
	Name:                "My API Check",
	Type:                checkly.TypeAPI,
	Frequency:           5,
	Activated:           true,
	Muted:               false,
	ShouldFail:          false,
	DoubleCheck:         false,
	SSLCheck:            true,
	SSLCheckDomain:      "example.com",
	LocalSetupScript:    "",
	LocalTearDownScript: "",
	Locations: []string{
		"eu-west-1",
		"ap-northeast-2",
	},
	Tags: []string{
		"foo",
		"bar",
	},
	AlertChannels:          alertChannels,
	AlertSettings:          alertSettings,
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
			checkly.Assertion{
				Source:     checkly.StatusCode,
				Comparison: checkly.Equals,
				Target:     "200",
			},
		},
		Body:     "",
		BodyType: "NONE",
	},
}

var browserCheck = checkly.Check{
	Name:           "My Browser Check",
	Type:           checkly.TypeBrowser,
	Frequency:      5,
	Activated:      true,
	Muted:          false,
	ShouldFail:     false,
	DoubleCheck:    false,
	SSLCheck:       true,
	SSLCheckDomain: "example.com",
	Locations:      []string{"eu-west-1"},
	AlertChannels:  alertChannels,
	AlertSettings:  alertSettings,
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
	Request: checkly.Request{
		Method: http.MethodGet,
		URL:    "http://example.com",
	},
}

func main() {
	apiKey := os.Getenv("CHECKLY_API_KEY")
	if apiKey == "" {
		log.Fatal("no CHECKLY_API_KEY set")
	}
	client := checkly.NewClient(apiKey)
	// uncomment this to enable dumping of API requests and responses
	// client.Debug = os.Stdout
	for _, check := range []checkly.Check{apiCheck, browserCheck} {
		ID, err := client.Create(check)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("New check created with ID %s\n", ID)
	}
}
