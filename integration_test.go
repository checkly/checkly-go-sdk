//go:build integration
// +build integration

package checkly_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/checkly/checkly-go-sdk/v2"
)

func setupClient(t *testing.T) checkly.Client {
	baseUrl := os.Getenv("CHECKLY_API_URL")
	if baseUrl == "" {
		baseUrl = "http://localhost:3000"
	}
	// WithAuthFromEnv requires both CHECKLY_API_KEY and CHECKLY_ACCOUNT_ID
	// and fails when either is not set.
	client, err := checkly.NewClient(
		checkly.WithBaseURL(baseUrl),
		checkly.WithAuthFromEnv(),
	)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

// TestGetCheckIntegration verifies that GetCheck reads a check created with a
// per-type method.
func TestGetCheckIntegration(t *testing.T) {
	ctx := context.TODO()
	client := setupClient(t)

	created, err := client.CreateAPICheck(ctx, checkly.APICheck{
		Name:                 "GetCheck integration check",
		Frequency:            10,
		Locations:            []string{"us-east-1"},
		Tags:                 []string{"integration"},
		DegradedResponseTime: 5000,
		MaxResponseTime:      20000,
		Request: checkly.Request{
			Method: "GET",
			URL:    "https://api.checklyhq.com",
		},
	})
	if err != nil {
		t.Fatalf("failed to create check: %v", err)
	}
	t.Cleanup(func() {
		_ = client.DeleteAPICheck(context.Background(), created.ID)
	})

	check, err := client.GetCheck(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get check: %v", err)
	}
	got, ok := check.(*checkly.APICheck)
	if !ok {
		t.Fatalf("got %T, want *checkly.APICheck", check)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %q, want %q", got.ID, created.ID)
	}
	if got.Name != created.Name {
		t.Errorf("Name = %q, want %q", got.Name, created.Name)
	}
	if got.Frequency != created.Frequency {
		t.Errorf("Frequency = %d, want %d", got.Frequency, created.Frequency)
	}
	if !cmp.Equal(created.Locations, got.Locations) {
		t.Error(cmp.Diff(created.Locations, got.Locations))
	}
	if !cmp.Equal(created.Tags, got.Tags) {
		t.Error(cmp.Diff(created.Tags, got.Tags))
	}
	if got.Request.URL != created.Request.URL {
		t.Errorf("Request.URL = %q, want %q", got.Request.URL, created.Request.URL)
	}
}

// assertGetCheckReturns creates a check in a subtest and verifies that
// GetCheck, whose request sets includeDependencies whatever the type, returns
// that check as a *T.
func assertGetCheckReturns[T any](t *testing.T, client checkly.Client, name string, create func() (*T, error)) {
	t.Run(name, func(t *testing.T) {
		ctx := context.TODO()
		created, err := create()
		if err != nil {
			t.Fatalf("failed to create %s check: %v", name, err)
		}
		id := reflect.ValueOf(created).Elem().FieldByName("ID").String()
		t.Cleanup(func() {
			_ = client.DeleteCheck(context.Background(), id)
		})

		check, err := client.GetCheck(ctx, id)
		if err != nil {
			t.Fatalf("failed to get %s check: %v", name, err)
		}
		got, ok := any(check).(*T)
		if !ok {
			t.Fatalf("got %T, want %T", check, created)
		}
		if gotID := reflect.ValueOf(got).Elem().FieldByName("ID").String(); gotID != id {
			t.Errorf("ID = %q, want %q", gotID, id)
		}
	})
}

// TestGetCheckConcreteTypesIntegration verifies that GetCheck returns the
// concrete type of every check type the SDK can create, apart from API checks,
// which TestGetCheckIntegration covers. Playwright checks need an uploaded code
// bundle, so only a unit test covers them.
func TestGetCheckConcreteTypesIntegration(t *testing.T) {
	ctx := context.TODO()
	client := setupClient(t)
	locations := []string{"us-east-1"}

	assertGetCheckReturns(t, client, "browser", func() (*checkly.BrowserCheck, error) {
		return client.CreateBrowserCheck(ctx, checkly.BrowserCheck{
			Name:      "GetCheck browser check",
			Frequency: 10,
			Locations: locations,
			Script:    "console.log('browser')",
		})
	})
	assertGetCheckReturns(t, client, "multi-step", func() (*checkly.MultiStepCheck, error) {
		return client.CreateMultiStepCheck(ctx, checkly.MultiStepCheck{
			Name:      "GetCheck multi-step check",
			Frequency: 10,
			Locations: locations,
			Script:    "console.log('multi-step')",
		})
	})
	assertGetCheckReturns(t, client, "heartbeat", func() (*checkly.HeartbeatMonitor, error) {
		return client.CreateHeartbeatMonitor(ctx, checkly.HeartbeatMonitor{
			Name: "GetCheck heartbeat monitor",
			AlertSettings: checkly.AlertSettings{
				EscalationType:     checkly.EscalationTypeRunBased,
				RunBasedEscalation: checkly.RunBasedEscalation{FailedRunThreshold: 1},
			},
			Heartbeat: checkly.Heartbeat{
				Period:     1,
				PeriodUnit: checkly.HeartbeatUnitDays,
				Grace:      1,
				GraceUnit:  checkly.HeartbeatUnitHours,
			},
		})
	})
	assertGetCheckReturns(t, client, "URL", func() (*checkly.URLMonitor, error) {
		return client.CreateURLMonitor(ctx, checkly.URLMonitor{
			Name:      "GetCheck URL monitor",
			Frequency: 10,
			Locations: locations,
			Request:   checkly.URLRequest{URL: "https://api.checklyhq.com"},
		})
	})
	assertGetCheckReturns(t, client, "TCP", func() (*checkly.TCPMonitor, error) {
		return client.CreateTCPMonitor(ctx, checkly.TCPMonitor{
			Name:      "GetCheck TCP monitor",
			Frequency: 10,
			Locations: locations,
			Request:   checkly.TCPRequest{Hostname: "api.checklyhq.com", Port: 443},
		})
	})
	assertGetCheckReturns(t, client, "DNS", func() (*checkly.DNSMonitor, error) {
		return client.CreateDNSMonitor(ctx, checkly.DNSMonitor{
			Name:      "GetCheck DNS monitor",
			Frequency: 10,
			Locations: locations,
			Request:   checkly.DNSRequest{RecordType: "A", Query: "api.checklyhq.com"},
		})
	})
	assertGetCheckReturns(t, client, "ICMP", func() (*checkly.ICMPMonitor, error) {
		return client.CreateICMPMonitor(ctx, checkly.ICMPMonitor{
			Name:                        "GetCheck ICMP monitor",
			Frequency:                   10,
			Locations:                   locations,
			DegradedPacketLossThreshold: 10,
			MaxPacketLossThreshold:      20,
			Request:                     checkly.ICMPRequest{Hostname: "api.checklyhq.com", PingCount: 5},
		})
	})
	assertGetCheckReturns(t, client, "gRPC", func() (*checkly.GRPCMonitor, error) {
		return client.CreateGRPCMonitor(ctx, checkly.GRPCMonitor{
			Name:      "GetCheck gRPC monitor",
			Frequency: 10,
			Locations: locations,
			Request: checkly.GRPCRequest{
				URL:        "grpc.example.com",
				Port:       443,
				Timeout:    30,
				GRPCConfig: checkly.GRPCConfig{Mode: "BEHAVIOR", Method: "example.Service/Method"},
			},
		})
	})
	assertGetCheckReturns(t, client, "SSL", func() (*checkly.SSLMonitor, error) {
		return client.CreateSSLMonitor(ctx, checkly.SSLMonitor{
			Name:      "GetCheck SSL monitor",
			Frequency: 10,
			Locations: locations,
			Request:   checkly.SSLRequest{SSLConfig: checkly.SSLConfig{Hostname: "api.checklyhq.com"}},
		})
	})
	assertGetCheckReturns(t, client, "traceroute", func() (*checkly.TracerouteMonitor, error) {
		return client.CreateTracerouteMonitor(ctx, checkly.TracerouteMonitor{
			Name:      "GetCheck traceroute monitor",
			Frequency: 10,
			Locations: locations,
			Request:   checkly.TracerouteRequest{URL: "api.checklyhq.com"},
		})
	})
}

// TestAPICheckRoundTripIntegration verifies that every field of a fully
// populated API check survives create, update and read: the API stores what
// the SDK sends and the SDK decodes what the API returns.
func TestAPICheckRoundTripIntegration(t *testing.T) {
	ctx := context.TODO()
	client := setupClient(t)

	description := "Round trip description"
	localSetup := "console.log('local setup')"
	localTearDown := "console.log('local teardown')"
	check := checkly.APICheck{
		Name:                 "Round trip check",
		Description:          &description,
		Frequency:            10,
		Activated:            true,
		Muted:                true,
		Locations:            []string{"us-east-1", "eu-west-1"},
		PrivateLocations:     &[]string{},
		Tags:                 []string{"integration", "round-trip"},
		DegradedResponseTime: 5000,
		MaxResponseTime:      20000,
		AlertSettings: &checkly.AlertSettings{
			EscalationType: checkly.EscalationTypeRunBased,
			RunBasedEscalation: checkly.RunBasedEscalation{
				FailedRunThreshold: 2,
			},
			TimeBasedEscalation: checkly.TimeBasedEscalation{
				MinutesFailingThreshold: 5,
			},
			Reminders: checkly.Reminders{
				Amount:   1,
				Interval: 5,
			},
			ParallelRunFailureThreshold: checkly.ParallelRunFailureThreshold{
				Percentage: 10,
			},
		},
		RetryStrategy: &checkly.RetryStrategy{
			Type:               "LINEAR",
			MaxRetries:         2,
			BaseBackoffSeconds: 60,
			MaxDurationSeconds: 600,
			SameRegion:         true,
		},
		LocalSetupScript:    &localSetup,
		LocalTearDownScript: &localTearDown,
		Request: checkly.Request{
			Method:          "GET",
			URL:             "https://api.checklyhq.com",
			FollowRedirects: true,
			BodyType:        checkly.BodyTypeNone,
			IPFamily:        checkly.IPFamilyIPv4,
			Headers: []checkly.KeyValue{
				{Key: "X-Test", Value: "foo"},
			},
			QueryParameters: []checkly.KeyValue{
				{Key: "query", Value: "param"},
			},
			Assertions: []checkly.Assertion{
				{
					Source:     checkly.AssertionSourceStatusCode,
					Comparison: checkly.AssertionComparisonEquals,
					Target:     "200",
				},
			},
		},
	}
	// The API assigns FrequencyOffset for a Frequency above 0.
	ignored := cmp.Options{ignoreAPICheckFields, cmpopts.IgnoreFields(checkly.APICheck{}, "FrequencyOffset"), cmpopts.EquateEmpty()}
	assertEqual := func(step string, got checkly.APICheck) {
		t.Helper()
		if diff := cmp.Diff(check, got, ignored); diff != "" {
			t.Errorf("%s: mismatch (-want +got):\n%s", step, diff)
		}
	}

	created, err := client.CreateAPICheck(ctx, check)
	if err != nil {
		t.Fatalf("failed to create check: %v", err)
	}
	t.Cleanup(func() {
		_ = client.DeleteAPICheck(context.Background(), created.ID)
	})
	assertEqual("create", *created)

	check.Name = "Round trip check updated"
	check.Tags = []string{"integration", "round-trip", "updated"}
	check.RetryStrategy.MaxRetries = 1
	check.Request.Headers = []checkly.KeyValue{{Key: "X-Test", Value: "bar"}}
	updated, err := client.UpdateAPICheck(ctx, created.ID, check)
	if err != nil {
		t.Fatalf("failed to update check: %v", err)
	}
	assertEqual("update", *updated)

	got, err := client.GetAPICheck(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get check: %v", err)
	}
	assertEqual("get", *got)
}

func TestCheckGroupUnset(t *testing.T) {
	ctx := context.TODO()

	client := setupClient(t)

	group, err := client.CreateGroupV2(ctx, checkly.GroupV2{
		Name:        "Test Group",
		Concurrency: 3,
		Tags:        []string{},
		AlertSettings: &checkly.AlertSettings{
			EscalationType: checkly.EscalationTypeRunBased,
			RunBasedEscalation: checkly.RunBasedEscalation{
				FailedRunThreshold: 1,
			},
		},
		Locations: []string{"us-east-1"},
	})

	if err != nil {
		t.Fatalf("failed to create group for check: %v", err)
	}

	defer func() {
		_ = client.DeleteGroupV2(ctx, group.ID)
	}()

	pendingCheck := checkly.APICheck{
		Name:                 "Foo check",
		Frequency:            10,
		DegradedResponseTime: 5000,
		MaxResponseTime:      20000,
		Request: checkly.Request{
			Method: "GET",
			URL:    "https://api.checklyhq.com",
		},
		AlertSettings: &checkly.AlertSettings{
			EscalationType: checkly.EscalationTypeRunBased,
			RunBasedEscalation: checkly.RunBasedEscalation{
				FailedRunThreshold: 1,
			},
		},
		Locations: []string{"us-east-1"},
		GroupID:   group.ID,
	}

	createdCheck, err := client.CreateAPICheck(ctx, pendingCheck)
	if err != nil {
		t.Fatalf("failed to create check: %v", err)
	}

	defer func() {
		_ = client.DeleteAPICheck(ctx, createdCheck.ID)
	}()

	if createdCheck.GroupID != group.ID {
		t.Fatalf("wrong GroupID after creation")
	}

	updateCheck := pendingCheck
	updateCheck.GroupID = 0

	if _, err := client.UpdateAPICheck(ctx, createdCheck.ID, updateCheck); err != nil {
		t.Fatalf("failed to update check: %v", err)
	}

	got, err := client.GetAPICheck(ctx, createdCheck.ID)
	if err != nil {
		t.Fatalf("failed to get check: %v", err)
	}

	if got.GroupID != 0 {
		t.Fatalf("wrong GroupID after update")
	}
}

// makeTestAlertChannel creates an email alert channel and deletes it when the
// test ends, after the test's deferred calls have deleted what subscribes to it.
func makeTestAlertChannel(t *testing.T, client checkly.Client) *checkly.AlertChannel {
	t.Helper()
	ac, err := client.CreateAlertChannel(
		context.Background(),
		checkly.AlertChannel{
			Type: checkly.AlertChannelTypeEmail,
			Email: &checkly.AlertChannelEmail{
				Address: "test@example.com",
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := client.DeleteAlertChannel(context.Background(), ac.ID); err != nil {
			t.Errorf("deleting alert channel %d: %v", ac.ID, err)
		}
	})
	return ac
}

func TestCreateGroupIntegration(t *testing.T) {
	client := setupClient(t)
	wantGroupCopy := wantGroup
	ac := makeTestAlertChannel(t, client)
	wantGroupCopy.AlertChannelSubscriptions[0].ChannelID = ac.ID
	gotGroup, err := client.CreateGroupV2(context.Background(), wantGroupCopy)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.DeleteGroupV2(context.Background(), gotGroup.ID) }()
	// These are set by the APIs
	ignored := cmpopts.IgnoreFields(checkly.GroupV2{}, "ID", "AlertChannelSubscriptions", "PrivateLocations", "CreatedAt", "UpdatedAt")
	if !cmp.Equal(wantGroupCopy, *gotGroup, ignored) {
		t.Fatal(cmp.Diff(wantGroupCopy, *gotGroup, ignored))
	}
}

func TestGetGroupIntegration(t *testing.T) {
	client := setupClient(t)
	ac := makeTestAlertChannel(t, client)
	wantGroupCopy := wantGroup
	wantGroupCopy.AlertChannelSubscriptions[0].ChannelID = ac.ID
	group, err := client.CreateGroupV2(context.Background(), wantGroupCopy)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.DeleteGroupV2(context.Background(), group.ID) }()
	gotGroup, err := client.GetGroupV2(context.Background(), group.ID)
	if err != nil {
		t.Fatal(err)
	}
	// These are set by the API
	wantGroupCopy.AlertChannelSubscriptions = gotGroup.AlertChannelSubscriptions
	if !cmp.Equal(wantGroupCopy, *gotGroup, ignoreGroupFields) {
		t.Fatal(cmp.Diff(wantGroupCopy, *gotGroup, ignoreGroupFields))
	}
}

//Dashboard

func TestCreateDashboardIntegration(t *testing.T) {
	client := setupClient(t)

	gotDashboard, err := client.CreateDashboard(context.Background(), testDashboard)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.DeleteDashboard(context.Background(), gotDashboard.DashboardID) }()
	if !cmp.Equal(testDashboard, *gotDashboard, ignoreDashboardFields) {
		t.Fatal(cmp.Diff(testDashboard, *gotDashboard, ignoreDashboardFields))
	}
}

func TestGetDashboardIntegration(t *testing.T) {
	client := setupClient(t)
	dash, err := client.CreateDashboard(context.Background(), testDashboard)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.DeleteDashboard(context.Background(), dash.DashboardID) }()
	gotDashboard, err := client.GetDashboard(context.Background(), dash.DashboardID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testDashboard, *gotDashboard, ignoreDashboardFields) {
		t.Fatal(cmp.Diff(testDashboard, *gotDashboard, ignoreDashboardFields))
	}
}

// wantDashboardDefaults fails the test unless the dashboard has the API
// defaults of the fields the SDK leaves out when unset.
func wantDashboardDefaults(t *testing.T, d *checkly.Dashboard) {
	t.Helper()
	if d.RefreshRate != 60 || d.PaginationRate != 60 || d.Paginate == nil || !*d.Paginate ||
		d.SortByStatus || d.ShowGroupNames == nil || !*d.ShowGroupNames {
		t.Fatalf("want refreshRate 60, paginationRate 60, paginate true, sortByStatus false, showGroupNames true; got %d, %d, %v, %v, %v",
			d.RefreshRate, d.PaginationRate, fmtBoolPtr(d.Paginate), d.SortByStatus, fmtBoolPtr(d.ShowGroupNames))
	}
}

func fmtBoolPtr(b *bool) string {
	if b == nil {
		return "nil"
	}
	return fmt.Sprint(*b)
}

func TestDashboardDefaultsIntegration(t *testing.T) {
	client := setupClient(t)
	ctx := context.Background()
	dashboard := checkly.Dashboard{
		CustomURL: fmt.Sprintf("sdk-defaults-%d", time.Now().UnixNano()),
		Header:    "Defaults",
	}

	created, err := client.CreateDashboard(ctx, dashboard)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.DeleteDashboard(ctx, created.DashboardID) }()
	wantDashboardDefaults(t, created)

	falseVal := false
	changed := dashboard
	changed.RefreshRate = 300
	changed.PaginationRate = 30
	changed.Paginate = &falseVal
	changed.SortByStatus = true
	changed.ShowGroupNames = &falseVal
	got, err := client.UpdateDashboard(ctx, created.DashboardID, changed)
	if err != nil {
		t.Fatal(err)
	}
	if got.RefreshRate != 300 || got.PaginationRate != 30 || got.Paginate == nil || *got.Paginate ||
		!got.SortByStatus || got.ShowGroupNames == nil || *got.ShowGroupNames {
		t.Fatalf("update did not set the fields: got %d, %d, %v, %v, %v",
			got.RefreshRate, got.PaginationRate, fmtBoolPtr(got.Paginate), got.SortByStatus, fmtBoolPtr(got.ShowGroupNames))
	}

	// The update endpoint applies the create defaults to omitted keys.
	if _, err := client.UpdateDashboard(ctx, created.DashboardID, dashboard); err != nil {
		t.Fatal(err)
	}
	read, err := client.GetDashboard(ctx, created.DashboardID)
	if err != nil {
		t.Fatal(err)
	}
	wantDashboardDefaults(t, read)
}

func TestDashboardClearTextsIntegration(t *testing.T) {
	client := setupClient(t)
	ctx := context.Background()
	dashboard := checkly.Dashboard{
		CustomURL:   fmt.Sprintf("sdk-clear-%d", time.Now().UnixNano()),
		Header:      "Header",
		Description: "Description",
		Logo:        "https://www.checklyhq.com/logo.svg",
		Favicon:     "https://www.checklyhq.com/favicon.ico",
		Link:        "https://www.checklyhq.com",
	}
	created, err := client.CreateDashboard(ctx, dashboard)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.DeleteDashboard(ctx, created.DashboardID) }()

	if _, err := client.UpdateDashboard(ctx, created.DashboardID, checkly.Dashboard{CustomURL: dashboard.CustomURL}); err != nil {
		t.Fatal(err)
	}
	read, err := client.GetDashboard(ctx, created.DashboardID)
	if err != nil {
		t.Fatal(err)
	}
	if read.Header != "" || read.Description != "" || read.Logo != "" || read.Favicon != "" || read.Link != "" {
		t.Fatalf("update with empty texts did not clear them: header %q, description %q, logo %q, favicon %q, link %q",
			read.Header, read.Description, read.Logo, read.Favicon, read.Link)
	}
}

//Maintenance Windows

func TestCreateMaintenanceWindowIntegration(t *testing.T) {
	client := setupClient(t)

	gotMaintenanceWindow, err := client.CreateMaintenanceWindow(context.Background(), testMaintenanceWindow)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.DeleteMaintenanceWindow(context.Background(), gotMaintenanceWindow.ID) }()
	if !cmp.Equal(testMaintenanceWindow, *gotMaintenanceWindow, ignoreMaintenanceWindowFields) {
		t.Fatal(cmp.Diff(testMaintenanceWindow, *gotMaintenanceWindow, ignoreMaintenanceWindowFields))
	}
}

func TestGetMaintenanceWindowIntegration(t *testing.T) {
	client := setupClient(t)
	dash, err := client.CreateMaintenanceWindow(context.Background(), testMaintenanceWindow)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.DeleteMaintenanceWindow(context.Background(), dash.ID) }()
	gotMaintenanceWindow, err := client.GetMaintenanceWindow(context.Background(), dash.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testMaintenanceWindow, *gotMaintenanceWindow, ignoreMaintenanceWindowFields) {
		t.Fatal(cmp.Diff(testMaintenanceWindow, *gotMaintenanceWindow, ignoreMaintenanceWindowFields))
	}
}

// PrivateLocations

func TestCreatePrivateLocationIntegration(t *testing.T) {
	client := setupClient(t)

	gotPrivateLocation, err := client.CreatePrivateLocation(context.Background(), testPrivateLocation)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.DeletePrivateLocation(context.Background(), gotPrivateLocation.ID) }()
	if !cmp.Equal(testPrivateLocation, *gotPrivateLocation, ignorePrivateLocationFields) {
		t.Fatal(cmp.Diff(testPrivateLocation, *gotPrivateLocation, ignorePrivateLocationFields))
	}
}
func TestGetPrivateLocationIntegration(t *testing.T) {
	client := setupClient(t)
	pl, err := client.CreatePrivateLocation(context.Background(), testPrivateLocation)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.DeletePrivateLocation(context.Background(), pl.ID) }()
	gotPrivateLocation, err := client.GetPrivateLocation(context.Background(), pl.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testPrivateLocation, *gotPrivateLocation, ignorePrivateLocationFields) {
		t.Fatal(cmp.Diff(testPrivateLocation, *gotPrivateLocation, ignorePrivateLocationFields))
	}
}

func TestTCPCheckCRUD(t *testing.T) {
	ctx := context.TODO()

	client := setupClient(t)

	pendingCheck := checkly.TCPMonitor{
		Name:      "TestTCPCheckCRUD",
		Muted:     false,
		Frequency: 1,
		Locations: []string{"eu-west-1"},
		Request: checkly.TCPRequest{
			Hostname: "api.checklyhq.com",
			Port:     443,
		},
	}

	createdCheck, err := client.CreateTCPMonitor(ctx, pendingCheck)
	if err != nil {
		t.Fatalf("failed to create TCP check: %v", err)
	}
	var didDelete bool
	defer func() {
		if !didDelete {
			_ = client.DeleteCheck(ctx, createdCheck.ID)
		}
	}()

	if createdCheck.Muted != false {
		t.Fatalf("expected Muted to be false after creation")
	}

	_, err = client.GetTCPMonitor(ctx, createdCheck.ID)
	if err != nil {
		t.Fatalf("failed to get TCP check: %v", err)
	}

	updateCheck := *createdCheck
	updateCheck.Muted = true

	updatedCheck, err := client.UpdateTCPMonitor(ctx, createdCheck.ID, updateCheck)
	if err != nil {
		t.Fatalf("failed to update TCP check: %v", err)
	}

	if updatedCheck.Muted != true {
		t.Fatalf("expected Muted to be true after update")
	}

	didDelete = true
	err = client.DeleteCheck(ctx, createdCheck.ID)
	if err != nil {
		t.Fatalf("failed to delete TCP check: %v", err)
	}
}

func TestTCPMonitorGroupUnset(t *testing.T) {
	ctx := context.TODO()

	client := setupClient(t)

	group, err := client.CreateGroupV2(ctx, checkly.GroupV2{
		Name:        "Test Group",
		Concurrency: 3,
		Tags:        []string{},
		AlertSettings: &checkly.AlertSettings{
			EscalationType: checkly.EscalationTypeRunBased,
			RunBasedEscalation: checkly.RunBasedEscalation{
				FailedRunThreshold: 1,
			},
		},
		Locations: []string{"us-east-1"},
	})

	if err != nil {
		t.Fatalf("failed to create group for TCP monitor: %v", err)
	}

	defer func() {
		_ = client.DeleteGroupV2(ctx, group.ID)
	}()

	pendingMonitor := checkly.TCPMonitor{
		Name:      "Foo monitor",
		Frequency: 1,
		Request: checkly.TCPRequest{
			Hostname: "api.checklyhq.com",
			Port:     443,
		},
		Locations: []string{"us-east-1"},
		GroupID:   group.ID,
	}

	createdMonitor, err := client.CreateTCPMonitor(ctx, pendingMonitor)
	if err != nil {
		t.Fatalf("failed to create TCP monitor: %v", err)
	}

	defer func() {
		_ = client.DeleteTCPMonitor(ctx, createdMonitor.ID)
	}()

	if createdMonitor.GroupID != group.ID {
		t.Fatalf("wrong GroupID after creation")
	}

	updateMonitor := pendingMonitor
	updateMonitor.GroupID = 0

	updatedMonitor, err := client.UpdateTCPMonitor(ctx, createdMonitor.ID, updateMonitor)
	if err != nil {
		t.Fatalf("failed to update TCP monitor: %v", err)
	}

	if updatedMonitor.GroupID != 0 {
		t.Fatalf("wrong GroupID after update")
	}
}

func TestClientCertificateCRD(t *testing.T) {
	ctx := context.TODO()

	client := setupClient(t)

	pendingClientCertificate := checkly.ClientCertificate{
		Host:        "*.acme.com",
		Certificate: "-----BEGIN CERTIFICATE-----\nMIICDzCCAbagAwIBAgIUMTZlfGA7WcD8e4/zt2MqxvEgQPYwCgYIKoZIzj0EAwIw\nVDELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAkNBMREwDwYDVQQHDAhUb29udG93bjES\nMBAGA1UECgwJQWNtZSBJbmMuMREwDwYDVQQDDAhhY21lLmNvbTAeFw0yNTAzMDMw\nNTQ2NTJaFw00OTEwMjMwNTQ2NTJaMHgxCzAJBgNVBAYTAlVTMQswCQYDVQQIDAJD\nQTERMA8GA1UEBwwIVG9vbnRvd24xEjAQBgNVBAoMCUFjbWUgSW5jLjEXMBUGA1UE\nAwwOV2lsZSBFLiBDb3lvdGUxHDAaBgkqhkiG9w0BCQEWDXdpbGVAYWNtZS5jb20w\nWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAATAjjDGsKFS1qgdNqziDZoD5hamTfdH\n0P+Ukk1RIue57QYVXhQSyNzcEz15kQnwYezEqfN+FtjtTwdk/CgnAELlo0IwQDAd\nBgNVHQ4EFgQU9C9CpZqM2WMrOs3vAYsc5GbjyzswHwYDVR0jBBgwFoAUnlOyzF/N\nK7YmKQegLdbdyIOCT/UwCgYIKoZIzj0EAwIDRwAwRAIgGgSnBymlH4MkZCVk5DYH\nPdnDo2Xf5uFi1Eyn2LTYP1MCIEtiGtsf0qYv6NzIPd5uTTZoB/8hPrAgM1QzWG4O\n3C/I\n-----END CERTIFICATE-----\n",
		PrivateKey:  "-----BEGIN ENCRYPTED PRIVATE KEY-----\nMIH0MF8GCSqGSIb3DQEFDTBSMDEGCSqGSIb3DQEFDDAkBBA5yR3aqy8mZD2wQzp1\nFH2JAgIIADAMBggqhkiG9w0CCQUAMB0GCWCGSAFlAwQBKgQQA49YCnXvfJ2CsQsV\n9C5JJwSBkNkWunSlqyeVW6OFa/+OjlLArgTGvW5ul08qu/145O9PO4Nr2CXeK5N2\nuvHwkWGfD8IVke+sgZPUjLoHsJ4h4AnyxlNHpIxgOfm0CoXT7PTaFb//d5NC6XyB\nK7ZpBzIThGlbuS/b9wp4MPmSaJn5Fci+84VG7KYK5RxU0fcU0rGSBynrZw803wnO\nFjP7qaq5bw==\n-----END ENCRYPTED PRIVATE KEY-----\n",
		Passphrase:  "secret password",
		TrustedCA:   "-----BEGIN CERTIFICATE-----\nMIIB/jCCAaOgAwIBAgIUZzxdNpoDYXaNiIBsh0/s++I+ZOEwCgYIKoZIzj0EAwIw\nVDELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAkNBMREwDwYDVQQHDAhUb29udG93bjES\nMBAGA1UECgwJQWNtZSBJbmMuMREwDwYDVQQDDAhhY21lLmNvbTAeFw0yNTAzMDMw\nNTQzMDZaFw0yNTA0MDIwNTQzMDZaMFQxCzAJBgNVBAYTAlVTMQswCQYDVQQIDAJD\nQTERMA8GA1UEBwwIVG9vbnRvd24xEjAQBgNVBAoMCUFjbWUgSW5jLjERMA8GA1UE\nAwwIYWNtZS5jb20wWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARDH3KGK6Vsk1A4\nyGf9ItQIS3yuAOi0n0ihmPzIOOOEN0c758ETABeUdgH55bakdx6q5KYSxf4TuXsJ\n2nCihqVVo1MwUTAdBgNVHQ4EFgQUnlOyzF/NK7YmKQegLdbdyIOCT/UwHwYDVR0j\nBBgwFoAUnlOyzF/NK7YmKQegLdbdyIOCT/UwDwYDVR0TAQH/BAUwAwEB/zAKBggq\nhkjOPQQDAgNJADBGAiEA/cJ9jV8MQz4ypQsFvUatrnbxyHO0f+pJhf09pAk6Kj8C\nIQCkSbope5r0KlVdqBeFF8wCfE3plwpelve3jqVIz6MedQ==\n-----END CERTIFICATE-----\n",
	}

	createdClientCertificate, err := client.CreateClientCertificate(ctx, pendingClientCertificate)
	if err != nil {
		t.Fatalf("failed to create client certificate: %v", err)
	}
	var didDelete bool
	defer func() {
		if !didDelete {
			_ = client.DeleteClientCertificate(ctx, createdClientCertificate.ID)
		}
	}()

	readClientCertificate, err := client.GetClientCertificate(ctx, createdClientCertificate.ID)
	if err != nil {
		t.Fatalf("failed to get client certificate: %v", err)
	}

	if !cmp.Equal(createdClientCertificate, readClientCertificate) {
		t.Fatal(cmp.Diff(createdClientCertificate, readClientCertificate, ignorePrivateLocationFields))
	}

	didDelete = true
	err = client.DeleteClientCertificate(ctx, createdClientCertificate.ID)
	if err != nil {
		t.Fatalf("failed to delete client certificate: %v", err)
	}
}

func TestStatusPageServiceCRUD(t *testing.T) {
	ctx := context.TODO()

	client := setupClient(t)

	pendingStatusPageService := checkly.StatusPageService{
		Name: "Foo service",
	}

	createdStatusPageService, err := client.CreateStatusPageService(ctx, pendingStatusPageService)
	if err != nil {
		t.Fatalf("failed to create status page service: %v", err)
	}
	var didDelete bool
	defer func() {
		if !didDelete {
			_ = client.DeleteStatusPageService(ctx, createdStatusPageService.ID)
		}
	}()

	readStatusPageService, err := client.GetStatusPageService(ctx, createdStatusPageService.ID)
	if err != nil {
		t.Fatalf("failed to get status page service: %v", err)
	}

	if !cmp.Equal(createdStatusPageService, readStatusPageService) {
		t.Fatal(cmp.Diff(createdStatusPageService, readStatusPageService, ignoreStatusPageFields))
	}

	updateStatusPageService := *createdStatusPageService
	updateStatusPageService.Name = "Bar service"

	updatedStatusPageService, err := client.UpdateStatusPageService(ctx, createdStatusPageService.ID, updateStatusPageService)
	if err != nil {
		t.Fatalf("failed to update status page service: %v", err)
	}

	if updatedStatusPageService.Name != "Bar service" {
		t.Fatalf("expected Name to change after update")
	}

	didDelete = true
	err = client.DeleteStatusPageService(ctx, createdStatusPageService.ID)
	if err != nil {
		t.Fatalf("failed to delete status page service: %v", err)
	}
}

func TestStatusPageCRUD(t *testing.T) {
	ctx := context.TODO()

	client := setupClient(t)

	pendingStatusPageService := checkly.StatusPageService{
		Name: "Foo service",
	}

	createdStatusPageService, err := client.CreateStatusPageService(ctx, pendingStatusPageService)
	if err != nil {
		t.Fatalf("failed to create status page service: %v", err)
	}
	defer func() {
		_ = client.DeleteStatusPageService(ctx, createdStatusPageService.ID)
	}()

	pendingStatusPage := checkly.StatusPage{
		Name: "Foo status page",
		URL:  "foo-status-page",
		Cards: []checkly.StatusPageCard{
			{
				Name: "Foo card",
				Services: []checkly.StatusPageService{
					{
						ID: createdStatusPageService.ID,
					},
				},
			},
		},
	}

	createdStatusPage, err := client.CreateStatusPage(ctx, pendingStatusPage)
	if err != nil {
		t.Fatalf("failed to create status page: %v", err)
	}
	var didDelete bool
	defer func() {
		if !didDelete {
			_ = client.DeleteStatusPage(ctx, createdStatusPage.ID)
		}
	}()

	readStatusPage, err := client.GetStatusPage(ctx, createdStatusPage.ID)
	if err != nil {
		t.Fatalf("failed to get status page: %v", err)
	}

	if len(readStatusPage.Cards) == 0 {
		t.Fatal("Expected status page to have cards")
	}

	if len(readStatusPage.Cards[0].Services) == 0 {
		t.Fatal("Expected status page card to have services")
	}

	// Fill out the name to allow cmp to work better.
	readStatusPage.Cards[0].Services[0].Name = createdStatusPageService.Name

	if !cmp.Equal(createdStatusPage, readStatusPage) {
		t.Fatal(cmp.Diff(createdStatusPage, readStatusPage, ignoreStatusPageFields))
	}

	updateStatusPage := *createdStatusPage
	updateStatusPage.Name = "Bar status page"

	updatedStatusPage, err := client.UpdateStatusPage(ctx, createdStatusPage.ID, updateStatusPage)
	if err != nil {
		t.Fatalf("failed to update status page: %v", err)
	}

	if updatedStatusPage.Name != "Bar status page" {
		t.Fatalf("expected Name to change after update")
	}

	didDelete = true
	err = client.DeleteStatusPage(ctx, createdStatusPage.ID)
	if err != nil {
		t.Fatalf("failed to delete status page: %v", err)
	}
}

func TestStatusPageV3CRUD(t *testing.T) {
	ctx := context.TODO()

	client := setupClient(t)

	pendingStatusPage := checkly.StatusPageV3{
		Name:          "Foo v3 status page",
		URL:           "foo-v3-status-page",
		Description:   "All Foo systems",
		SupportLink:   "mailto:support@example.org",
		FooterText:    "Foo Inc.",
		AllowIndexing: true,
	}

	createdStatusPage, err := client.CreateStatusPageV3(ctx, pendingStatusPage)
	if err != nil {
		t.Fatalf("failed to create v3 status page: %v", err)
	}
	var didDelete bool
	defer func() {
		if !didDelete {
			_ = client.DeleteStatusPageV3(ctx, createdStatusPage.ID)
		}
	}()

	readStatusPage, err := client.GetStatusPageV3(ctx, createdStatusPage.ID)
	if err != nil {
		t.Fatalf("failed to get v3 status page: %v", err)
	}
	if !cmp.Equal(createdStatusPage, readStatusPage) {
		t.Fatal(cmp.Diff(createdStatusPage, readStatusPage))
	}
	// Reads always carry a complete palette, the defaults when none is set.
	if readStatusPage.ThemeColors == nil || readStatusPage.ThemeColors.Light.LinkFontColor == "" {
		t.Fatalf("expected a complete default theme colors palette, got %+v", readStatusPage.ThemeColors)
	}

	updateStatusPage := *createdStatusPage
	updateStatusPage.Name = "Bar v3 status page"
	// The update endpoint only touches fields present in the payload; the
	// optional fields serialize even when empty, so zeroing them here must
	// clear the values set on create.
	updateStatusPage.Description = ""
	updateStatusPage.SupportLink = ""
	updateStatusPage.FooterText = ""

	updatedStatusPage, err := client.UpdateStatusPageV3(ctx, createdStatusPage.ID, updateStatusPage)
	if err != nil {
		t.Fatalf("failed to update v3 status page: %v", err)
	}
	if updatedStatusPage.Name != "Bar v3 status page" {
		t.Fatalf("expected Name to change after update")
	}
	if updatedStatusPage.Description != "" || updatedStatusPage.SupportLink != "" || updatedStatusPage.FooterText != "" {
		t.Fatalf("expected cleared optional fields to be unset, got description %q, supportLink %q, footerText %q",
			updatedStatusPage.Description, updatedStatusPage.SupportLink, updatedStatusPage.FooterText)
	}

	// A partial configuration is filled up with the type's defaults.
	expandedByDefault := true
	group, err := client.CreateStatusPageComponentV3(ctx, createdStatusPage.ID, checkly.StatusPageComponentV3{
		Type:         checkly.StatusPageComponentV3TypeGroup,
		Name:         "Foo group",
		DisplayOrder: 0,
		Configuration: &checkly.StatusPageComponentV3Configuration{
			ExpandedByDefault: &expandedByDefault,
		},
	})
	if err != nil {
		t.Fatalf("failed to create GROUP component: %v", err)
	}
	if group.Configuration == nil ||
		group.Configuration.ExpandedByDefault == nil || !*group.Configuration.ExpandedByDefault ||
		group.Configuration.ShowHistoricalData == nil || !*group.Configuration.ShowHistoricalData {
		t.Fatalf("expected GROUP configuration {expandedByDefault: true, showHistoricalData: true}, got %+v", group.Configuration)
	}

	pendingComponent := checkly.StatusPageComponentV3{
		Name:         "Foo API",
		DisplayOrder: 1,
		ParentID:     group.ID,
	}

	createdComponent, err := client.CreateStatusPageComponentV3(ctx, createdStatusPage.ID, pendingComponent)
	if err != nil {
		t.Fatalf("failed to create SERVICE component: %v", err)
	}
	if createdComponent.Type != checkly.StatusPageComponentV3TypeService {
		t.Fatalf("expected component type to default to SERVICE, got %q", createdComponent.Type)
	}
	if createdComponent.ParentID != group.ID {
		t.Fatalf("expected component to sit under the group")
	}
	// A component created without a configuration carries its type's
	// defaults.
	if createdComponent.Configuration == nil ||
		createdComponent.Configuration.ShowHistoricalData == nil || !*createdComponent.Configuration.ShowHistoricalData {
		t.Fatalf("expected SERVICE configuration {showHistoricalData: true}, got %+v", createdComponent.Configuration)
	}

	readComponent, err := client.GetStatusPageComponentV3(ctx, createdStatusPage.ID, createdComponent.ID)
	if err != nil {
		t.Fatalf("failed to get component: %v", err)
	}
	if !cmp.Equal(createdComponent, readComponent) {
		t.Fatal(cmp.Diff(createdComponent, readComponent))
	}

	updateComponent := *createdComponent
	updateComponent.Name = "Bar API"

	updatedComponent, err := client.UpdateStatusPageComponentV3(ctx, createdStatusPage.ID, createdComponent.ID, updateComponent)
	if err != nil {
		t.Fatalf("failed to update component: %v", err)
	}
	if updatedComponent.Name != "Bar API" {
		t.Fatalf("expected component Name to change after update")
	}

	pendingRule := checkly.StatusPageAutomationRuleV3{
		Name:                  "Foo API outage",
		Enabled:               true,
		FirstUpdate:           "We are investigating an issue.",
		LastUpdate:            "The issue has been resolved.",
		NotifySubscribers:     true,
		CoolDownWindowMinutes: 5,
		Tags:                  []string{"foo-api"},
		Components: []checkly.StatusPageAutomationRuleComponentV3{
			{
				ComponentID:  createdComponent.ID,
				TargetImpact: checkly.StatusPageTargetImpactV3MajorOutage,
			},
		},
	}

	createdRule, err := client.CreateStatusPageAutomationRuleV3(ctx, createdStatusPage.ID, pendingRule)
	if err != nil {
		t.Fatalf("failed to create automation rule: %v", err)
	}

	readRule, err := client.GetStatusPageAutomationRuleV3(ctx, createdStatusPage.ID, createdRule.ID)
	if err != nil {
		t.Fatalf("failed to get automation rule: %v", err)
	}
	if !cmp.Equal(createdRule, readRule) {
		t.Fatal(cmp.Diff(createdRule, readRule))
	}

	updateRule := *createdRule
	updateRule.Name = "Bar API outage"

	updatedRule, err := client.UpdateStatusPageAutomationRuleV3(ctx, createdStatusPage.ID, createdRule.ID, updateRule)
	if err != nil {
		t.Fatalf("failed to update automation rule: %v", err)
	}
	if updatedRule.Name != "Bar API outage" {
		t.Fatalf("expected rule Name to change after update")
	}

	err = client.DeleteStatusPageAutomationRuleV3(ctx, createdStatusPage.ID, createdRule.ID)
	if err != nil {
		t.Fatalf("failed to delete automation rule: %v", err)
	}
	// A group must never be empty, so its last member cannot be deleted (or
	// detached) directly: delete the group first, which detaches its
	// children, then the component.
	err = client.DeleteStatusPageComponentV3(ctx, createdStatusPage.ID, group.ID)
	if err != nil {
		t.Fatalf("failed to delete group component: %v", err)
	}
	err = client.DeleteStatusPageComponentV3(ctx, createdStatusPage.ID, createdComponent.ID)
	if err != nil {
		t.Fatalf("failed to delete component: %v", err)
	}

	didDelete = true
	err = client.DeleteStatusPageV3(ctx, createdStatusPage.ID)
	if err != nil {
		t.Fatalf("failed to delete v3 status page: %v", err)
	}
}

func TestURLMonitorCRUD(t *testing.T) {
	ctx := context.TODO()

	client := setupClient(t)

	pendingMonitor := checkly.URLMonitor{
		Name:      "Foo monitor",
		Frequency: 1,
		Request: checkly.URLRequest{
			URL: "https://welcome.checklyhq.com/foo",
			Assertions: []checkly.Assertion{
				{
					Source:     "STATUS_CODE",
					Target:     "200",
					Comparison: "EQUALS",
				},
			},
		},
		Locations: []string{"us-east-1"},
	}

	createdMonitor, err := client.CreateURLMonitor(ctx, pendingMonitor)
	if err != nil {
		t.Fatalf("failed to create URL monitor: %v", err)
	}

	var deleteOnce sync.Once
	defer func() {
		deleteOnce.Do(func() {
			_ = client.DeleteURLMonitor(ctx, createdMonitor.ID)
		})
	}()

	readMonitor, err := client.GetURLMonitor(ctx, createdMonitor.ID)
	if err != nil {
		t.Fatalf("failed to get URL monitor: %v", err)
	}

	if readMonitor.Name != "Foo monitor" {
		t.Fatalf("wrong Name after creation")
	}

	if readMonitor.Request.URL != "https://welcome.checklyhq.com/foo" {
		t.Fatalf("wrong URL after creation")
	}

	updateMonitor := checkly.URLMonitor{
		Name:      "Bar monitor",
		Frequency: 1,
		Request: checkly.URLRequest{
			URL: "https://welcome.checklyhq.com/bar",
			Assertions: []checkly.Assertion{
				{
					Source:     "STATUS_CODE",
					Target:     "404",
					Comparison: "EQUALS",
				},
			},
		},
	}

	updatedMonitor, err := client.UpdateURLMonitor(ctx, createdMonitor.ID, updateMonitor)
	if err != nil {
		t.Fatalf("failed to update URL monitor: %v", err)
	}

	if updatedMonitor.Name != "Bar monitor" {
		t.Fatalf("wrong Name after update")
	}

	if updatedMonitor.Request.URL != "https://welcome.checklyhq.com/bar" {
		t.Fatalf("wrong URL after update")
	}

	if len(updatedMonitor.Request.Assertions) != 1 {
		t.Fatalf("wrong Assertion count after update")
	}

	if updatedMonitor.Request.Assertions[0].Target != "404" {
		t.Fatalf("wrong Assertion.Target after update")
	}

	deleteOnce.Do(func() {
		err := client.DeleteURLMonitor(ctx, createdMonitor.ID)
		if err != nil {
			t.Fatalf("failed to delete URL monitor: %v", err)
		}
	})
}

func TestURLMonitorGroupUnset(t *testing.T) {
	ctx := context.TODO()

	client := setupClient(t)

	group, err := client.CreateGroupV2(ctx, checkly.GroupV2{
		Name:        "Test Group",
		Concurrency: 3,
		Tags:        []string{},
		AlertSettings: &checkly.AlertSettings{
			EscalationType: checkly.EscalationTypeRunBased,
			RunBasedEscalation: checkly.RunBasedEscalation{
				FailedRunThreshold: 1,
			},
		},
		Locations: []string{"us-east-1"},
	})

	if err != nil {
		t.Fatalf("failed to create group for URL monitor: %v", err)
	}

	defer func() {
		_ = client.DeleteGroupV2(ctx, group.ID)
	}()

	pendingMonitor := checkly.URLMonitor{
		Name:      "Foo monitor",
		Frequency: 1,
		Request: checkly.URLRequest{
			URL:        "https://welcome.checklyhq.com/foo",
			Assertions: []checkly.Assertion{},
		},
		Locations: []string{"us-east-1"},
		GroupID:   group.ID,
	}

	createdMonitor, err := client.CreateURLMonitor(ctx, pendingMonitor)
	if err != nil {
		t.Fatalf("failed to create URL monitor: %v", err)
	}

	defer func() {
		_ = client.DeleteURLMonitor(ctx, createdMonitor.ID)
	}()

	if createdMonitor.GroupID != group.ID {
		t.Fatalf("wrong GroupID after creation")
	}

	updateMonitor := pendingMonitor
	updateMonitor.GroupID = 0

	updatedMonitor, err := client.UpdateURLMonitor(ctx, createdMonitor.ID, updateMonitor)
	if err != nil {
		t.Fatalf("failed to update URL monitor: %v", err)
	}

	if updatedMonitor.GroupID != 0 {
		t.Fatalf("wrong GroupID after update")
	}
}

func TestDNSMonitorCRUD(t *testing.T) {
	ctx := context.TODO()

	client := setupClient(t)

	pendingMonitor := checkly.DNSMonitor{
		Name:      "Foo monitor",
		Frequency: 1,
		Request: checkly.DNSRequest{
			RecordType: "A",
			Query:      "welcome.checklyhq.com",
			Assertions: []checkly.Assertion{
				{
					Source:     "RESPONSE_CODE",
					Target:     "NOERROR",
					Comparison: "EQUALS",
				},
			},
		},
		Locations: []string{"us-east-1"},
	}

	createdMonitor, err := client.CreateDNSMonitor(ctx, pendingMonitor)
	if err != nil {
		t.Fatalf("failed to create DNS monitor: %v", err)
	}

	var deleteOnce sync.Once
	defer func() {
		deleteOnce.Do(func() {
			_ = client.DeleteDNSMonitor(ctx, createdMonitor.ID)
		})
	}()

	readMonitor, err := client.GetDNSMonitor(ctx, createdMonitor.ID)
	if err != nil {
		t.Fatalf("failed to get DNS monitor: %v", err)
	}

	if readMonitor.Name != "Foo monitor" {
		t.Fatalf("wrong Name after creation")
	}

	if readMonitor.Request.Query != "welcome.checklyhq.com" {
		t.Fatalf("wrong Query after creation")
	}

	if len(readMonitor.Request.Assertions) != 1 {
		t.Fatalf("wrong Assertion count after creation")
	}

	if readMonitor.Request.Assertions[0].Target != "NOERROR" {
		t.Fatalf("wrong Assertion.Target after creation")
	}

	if len(readMonitor.Locations) != 1 {
		t.Fatalf("wrong Locations count after creation")
	}

	if readMonitor.Locations[0] != "us-east-1" {
		t.Fatalf("wrong Location after creation")
	}

	updateMonitor := checkly.DNSMonitor{
		Name:      "Bar monitor",
		Frequency: 1,
		Request: checkly.DNSRequest{
			RecordType: "A",
			Query:      "welcome2.checklyhq.com",
			Assertions: []checkly.Assertion{
				{
					Source:     "RESPONSE_CODE",
					Target:     "NXDOMAIN",
					Comparison: "EQUALS",
				},
			},
		},
		Locations: []string{"us-west-1"},
	}

	updatedMonitor, err := client.UpdateDNSMonitor(ctx, createdMonitor.ID, updateMonitor)
	if err != nil {
		t.Fatalf("failed to update DNS monitor: %v", err)
	}

	if updatedMonitor.Name != "Bar monitor" {
		t.Fatalf("wrong Name after update")
	}

	if updatedMonitor.Request.Query != "welcome2.checklyhq.com" {
		t.Fatalf("wrong Query after update")
	}

	if len(updatedMonitor.Request.Assertions) != 1 {
		t.Fatalf("wrong Assertion count after update")
	}

	if updatedMonitor.Request.Assertions[0].Target != "NXDOMAIN" {
		t.Fatalf("wrong Assertion.Target after update")
	}

	if len(updatedMonitor.Locations) != 1 {
		t.Fatalf("wrong Locations count after update")
	}

	if updatedMonitor.Locations[0] != "us-west-1" {
		t.Fatalf("wrong Location after update")
	}

	deleteOnce.Do(func() {
		err := client.DeleteDNSMonitor(ctx, createdMonitor.ID)
		if err != nil {
			t.Fatalf("failed to delete DNS monitor: %v", err)
		}
	})
}

func TestICMPMonitorCRUD(t *testing.T) {
	ctx := context.TODO()

	client := setupClient(t)

	pendingMonitor := checkly.ICMPMonitor{
		Name:                        "Foo ICMP monitor",
		Frequency:                   1,
		DegradedPacketLossThreshold: 10,
		MaxPacketLossThreshold:      20,
		Request: checkly.ICMPRequest{
			Hostname:  "welcome.checklyhq.com",
			IPFamily:  "IPv4",
			PingCount: 5,
			Assertions: []checkly.Assertion{
				{
					Source:     "LATENCY",
					Property:   "avg",
					Target:     "500",
					Comparison: "LESS_THAN",
				},
			},
		},
		Locations: []string{"us-east-1"},
	}

	createdMonitor, err := client.CreateICMPMonitor(ctx, pendingMonitor)
	if err != nil {
		t.Fatalf("failed to create ICMP monitor: %v", err)
	}

	var deleteOnce sync.Once
	defer func() {
		deleteOnce.Do(func() {
			_ = client.DeleteICMPMonitor(ctx, createdMonitor.ID)
		})
	}()

	readMonitor, err := client.GetICMPMonitor(ctx, createdMonitor.ID)
	if err != nil {
		t.Fatalf("failed to get ICMP monitor: %v", err)
	}

	if readMonitor.Name != "Foo ICMP monitor" {
		t.Fatalf("wrong Name after creation")
	}

	if readMonitor.Request.Hostname != "welcome.checklyhq.com" {
		t.Fatalf("wrong Hostname after creation")
	}

	if readMonitor.Request.PingCount != 5 {
		t.Fatalf("wrong PingCount after creation, got %d", readMonitor.Request.PingCount)
	}

	if len(readMonitor.Request.Assertions) != 1 {
		t.Fatalf("wrong Assertion count after creation")
	}

	if readMonitor.Request.Assertions[0].Source != "LATENCY" {
		t.Fatalf("wrong Assertion.Source after creation, got %s", readMonitor.Request.Assertions[0].Source)
	}

	if readMonitor.Request.Assertions[0].Property != "avg" {
		t.Fatalf("wrong Assertion.Property after creation, got %s", readMonitor.Request.Assertions[0].Property)
	}

	if readMonitor.Request.Assertions[0].Target != "500" {
		t.Fatalf("wrong Assertion.Target after creation")
	}

	if len(readMonitor.Locations) != 1 {
		t.Fatalf("wrong Locations count after creation")
	}

	if readMonitor.Locations[0] != "us-east-1" {
		t.Fatalf("wrong Location after creation")
	}

	updateMonitor := checkly.ICMPMonitor{
		Name:                        "Bar ICMP monitor",
		Frequency:                   1,
		DegradedPacketLossThreshold: 10,
		MaxPacketLossThreshold:      20,
		Request: checkly.ICMPRequest{
			Hostname:  "api.checklyhq.com",
			IPFamily:  "IPv4",
			PingCount: 3,
			Assertions: []checkly.Assertion{
				{
					Source:     "LATENCY",
					Property:   "max",
					Target:     "200",
					Comparison: "LESS_THAN",
				},
			},
		},
		Locations: []string{"us-west-1"},
	}

	updatedMonitor, err := client.UpdateICMPMonitor(ctx, createdMonitor.ID, updateMonitor)
	if err != nil {
		t.Fatalf("failed to update ICMP monitor: %v", err)
	}

	if updatedMonitor.Name != "Bar ICMP monitor" {
		t.Fatalf("wrong Name after update")
	}

	if updatedMonitor.Request.Hostname != "api.checklyhq.com" {
		t.Fatalf("wrong Hostname after update")
	}

	if updatedMonitor.Request.PingCount != 3 {
		t.Fatalf("wrong PingCount after update, got %d", updatedMonitor.Request.PingCount)
	}

	if len(updatedMonitor.Request.Assertions) != 1 {
		t.Fatalf("wrong Assertion count after update")
	}

	if updatedMonitor.Request.Assertions[0].Property != "max" {
		t.Fatalf("wrong Assertion.Property after update, got %s", updatedMonitor.Request.Assertions[0].Property)
	}

	if updatedMonitor.Request.Assertions[0].Target != "200" {
		t.Fatalf("wrong Assertion.Target after update")
	}

	if len(updatedMonitor.Locations) != 1 {
		t.Fatalf("wrong Locations count after update")
	}

	if updatedMonitor.Locations[0] != "us-west-1" {
		t.Fatalf("wrong Location after update")
	}

	deleteOnce.Do(func() {
		err := client.DeleteICMPMonitor(ctx, createdMonitor.ID)
		if err != nil {
			t.Fatalf("failed to delete ICMP monitor: %v", err)
		}
	})
}

func TestICMPMonitorGroupUnset(t *testing.T) {
	ctx := context.TODO()

	client := setupClient(t)

	group, err := client.CreateGroupV2(ctx, checkly.GroupV2{
		Name:        "Test Group",
		Concurrency: 3,
		Tags:        []string{},
		AlertSettings: &checkly.AlertSettings{
			EscalationType: checkly.EscalationTypeRunBased,
			RunBasedEscalation: checkly.RunBasedEscalation{
				FailedRunThreshold: 1,
			},
		},
		Locations: []string{"us-east-1"},
	})

	if err != nil {
		t.Fatalf("failed to create group for ICMP monitor: %v", err)
	}

	defer func() {
		_ = client.DeleteGroupV2(ctx, group.ID)
	}()

	pendingMonitor := checkly.ICMPMonitor{
		Name:                        "Foo ICMP monitor",
		Frequency:                   1,
		DegradedPacketLossThreshold: 10,
		MaxPacketLossThreshold:      20,
		Request: checkly.ICMPRequest{
			Hostname:   "welcome.checklyhq.com",
			Assertions: []checkly.Assertion{},
		},
		Locations: []string{"us-east-1"},
		GroupID:   group.ID,
	}

	createdMonitor, err := client.CreateICMPMonitor(ctx, pendingMonitor)
	if err != nil {
		t.Fatalf("failed to create ICMP monitor: %v", err)
	}

	defer func() {
		_ = client.DeleteICMPMonitor(ctx, createdMonitor.ID)
	}()

	if createdMonitor.GroupID != group.ID {
		t.Fatalf("wrong GroupID after creation")
	}

	updateMonitor := pendingMonitor
	updateMonitor.GroupID = 0

	updatedMonitor, err := client.UpdateICMPMonitor(ctx, createdMonitor.ID, updateMonitor)
	if err != nil {
		t.Fatalf("failed to update ICMP monitor: %v", err)
	}

	if updatedMonitor.GroupID != 0 {
		t.Fatalf("wrong GroupID after update")
	}
}

func TestDNSMonitorGroupUnset(t *testing.T) {
	ctx := context.TODO()

	client := setupClient(t)

	group, err := client.CreateGroupV2(ctx, checkly.GroupV2{
		Name:        "Test Group",
		Concurrency: 3,
		Tags:        []string{},
		AlertSettings: &checkly.AlertSettings{
			EscalationType: checkly.EscalationTypeRunBased,
			RunBasedEscalation: checkly.RunBasedEscalation{
				FailedRunThreshold: 1,
			},
		},
		Locations: []string{"us-east-1"},
	})

	if err != nil {
		t.Fatalf("failed to create group for DNS monitor: %v", err)
	}

	defer func() {
		_ = client.DeleteGroupV2(ctx, group.ID)
	}()

	pendingMonitor := checkly.DNSMonitor{
		Name:      "Foo monitor",
		Frequency: 1,
		Request: checkly.DNSRequest{
			RecordType: "A",
			Query:      "welcome.checklyhq.com",
			Assertions: []checkly.Assertion{},
		},
		Locations: []string{"us-east-1"},
		GroupID:   group.ID,
	}

	createdMonitor, err := client.CreateDNSMonitor(ctx, pendingMonitor)
	if err != nil {
		t.Fatalf("failed to create DNS monitor: %v", err)
	}

	defer func() {
		_ = client.DeleteDNSMonitor(ctx, createdMonitor.ID)
	}()

	if createdMonitor.GroupID != group.ID {
		t.Fatalf("wrong GroupID after creation")
	}

	updateMonitor := pendingMonitor
	updateMonitor.GroupID = 0

	updatedMonitor, err := client.UpdateDNSMonitor(ctx, createdMonitor.ID, updateMonitor)
	if err != nil {
		t.Fatalf("failed to update DNS monitor: %v", err)
	}

	if updatedMonitor.GroupID != 0 {
		t.Fatalf("wrong GroupID after update")
	}
}

// The alert channel update resets an omitted autoSubscribe to false.
func TestAlertChannelAutoSubscribeIntegration(t *testing.T) {
	ctx := context.TODO()
	client := setupClient(t)

	on := true
	pending := checkly.AlertChannel{
		Type:          checkly.AlertChannelTypeEmail,
		Email:         &checkly.AlertChannelEmail{Address: "test@example.com"},
		AutoSubscribe: &on,
	}
	created, err := client.CreateAlertChannel(ctx, pending)
	if err != nil {
		t.Fatalf("failed to create alert channel: %v", err)
	}
	defer func() {
		_ = client.DeleteAlertChannel(ctx, created.ID)
	}()
	if created.AutoSubscribe == nil || !*created.AutoSubscribe {
		t.Fatalf("expected autoSubscribe true after create, got %v", created.AutoSubscribe)
	}
	read, err := client.GetAlertChannel(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get alert channel: %v", err)
	}
	if read.AutoSubscribe == nil || !*read.AutoSubscribe {
		t.Fatalf("expected autoSubscribe true after read, got %v", read.AutoSubscribe)
	}

	pending.AutoSubscribe = nil
	updated, err := client.UpdateAlertChannel(ctx, created.ID, pending)
	if err != nil {
		t.Fatalf("failed to update alert channel: %v", err)
	}
	if updated.AutoSubscribe == nil || *updated.AutoSubscribe {
		t.Fatalf("expected autoSubscribe false after an update without it, got %v", updated.AutoSubscribe)
	}
}

// The intent, the auto-repair override and assertion regexes round-trip,
// and an update without the intent and override clears them.
func TestAPICheckSelfHealingFieldsIntegration(t *testing.T) {
	ctx := context.TODO()
	client := setupClient(t)

	off := false
	intent := &checkly.CheckIntent{
		Goal:             "The public API answers",
		RequiredOutcomes: []string{"The status endpoint returns 200"},
		MustPreserve:     []string{},
	}
	pending := checkly.APICheck{
		Name:      "Foo self-healing check",
		Frequency: 60,
		Locations: []string{"us-east-1"},
		Request: checkly.Request{
			Method:          "GET",
			URL:             "https://api.checklyhq.com/v1/status",
			Headers:         []checkly.KeyValue{},
			QueryParameters: []checkly.KeyValue{},
			Assertions: []checkly.Assertion{
				{Source: checkly.AssertionSourceStatusCode, Comparison: checkly.AssertionComparisonEquals, Target: "200"},
				{Source: checkly.AssertionSourceHeaders, Property: "content-type", Comparison: checkly.AssertionComparisonContains, Target: "json", Regex: `^(\w+/\w+)`},
			},
		},
		Intent:              intent,
		AIAutoRepairEnabled: &off,
	}
	created, err := client.CreateAPICheck(ctx, pending)
	if err != nil {
		t.Fatalf("failed to create API check: %v", err)
	}
	defer func() {
		_ = client.DeleteAPICheck(ctx, created.ID)
	}()

	read, err := client.GetAPICheck(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get API check: %v", err)
	}
	if !cmp.Equal(intent, read.Intent) {
		t.Fatalf("intent after create: %s", cmp.Diff(intent, read.Intent))
	}
	if read.AIAutoRepairEnabled == nil || *read.AIAutoRepairEnabled {
		t.Fatalf("expected aiAutoRepairEnabled false after create, got %v", read.AIAutoRepairEnabled)
	}
	if got := read.Request.Assertions; len(got) != 2 || got[0].Regex != "" || got[1].Regex != `^(\w+/\w+)` {
		t.Fatalf("expected the assertion regexes to round-trip, got %+v", got)
	}

	pending.Intent = nil
	pending.AIAutoRepairEnabled = nil
	updated, err := client.UpdateAPICheck(ctx, created.ID, pending)
	if err != nil {
		t.Fatalf("failed to update API check: %v", err)
	}
	if updated.Intent != nil || updated.AIAutoRepairEnabled != nil {
		t.Fatalf("expected intent and aiAutoRepairEnabled cleared after update, got %+v, %v", updated.Intent, updated.AIAutoRepairEnabled)
	}
	if got := updated.Request.Assertions; len(got) != 2 || got[1].Regex != `^(\w+/\w+)` {
		t.Fatalf("expected the assertion regex to survive the update, got %+v", got)
	}
}

// The proxy URL round-trips, and an update with an empty one removes it.
func TestPrivateLocationProxyURLIntegration(t *testing.T) {
	ctx := context.TODO()
	client := setupClient(t)

	proxy := "http://proxy.example.com:3128"
	pending := checkly.PrivateLocation{
		Name:     "Foo proxy location",
		SlugName: "foo-proxy-location",
		Icon:     "location",
		ProxyURL: &proxy,
	}
	created, err := client.CreatePrivateLocation(ctx, pending)
	if err != nil {
		t.Fatalf("failed to create private location: %v", err)
	}
	defer func() {
		_ = client.DeletePrivateLocation(ctx, created.ID)
	}()
	read, err := client.GetPrivateLocation(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get private location: %v", err)
	}
	if read.ProxyURL == nil || *read.ProxyURL != proxy {
		t.Fatalf("expected the proxy URL after create, got %v", read.ProxyURL)
	}

	empty := ""
	pending.ProxyURL = &empty
	if _, err := client.UpdatePrivateLocation(ctx, created.ID, pending); err != nil {
		t.Fatalf("failed to update private location: %v", err)
	}
	read, err = client.GetPrivateLocation(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get private location: %v", err)
	}
	if read.ProxyURL != nil {
		t.Fatalf("expected no proxy URL after update, got %v", read.ProxyURL)
	}
}

// A maintenance window links a status page with components through its
// component configs; a window read back can be sent unchanged, and an update
// with an empty list removes them.
func TestMaintenanceWindowComponentConfigsIntegration(t *testing.T) {
	ctx := context.TODO()
	client := setupClient(t)

	page, err := client.CreateStatusPageV3(ctx, checkly.StatusPageV3{
		Name: "Foo maintenance status page",
		URL:  "foo-maintenance-status-page",
	})
	if err != nil {
		t.Fatalf("failed to create v3 status page: %v", err)
	}
	defer func() {
		_ = client.DeleteStatusPageV3(ctx, page.ID)
	}()
	component, err := client.CreateStatusPageComponentV3(ctx, page.ID, checkly.StatusPageComponentV3{Name: "Foo API"})
	if err != nil {
		t.Fatalf("failed to create status page component: %v", err)
	}

	configs := []checkly.MaintenanceWindowComponentConfig{{
		StatusPageID: page.ID,
		Components: []checkly.MaintenanceWindowComponentImpact{
			{ComponentID: component.ID, Status: "PARTIAL_OUTAGE"},
		},
	}}
	pending := checkly.MaintenanceWindow{
		Name:     "Foo maintenance",
		StartsAt: "2030-01-01T00:00:00.000Z",
		EndsAt:   "2030-01-01T01:00:00.000Z",
		StatusPageVisibility: checkly.MaintenanceWindowStatusPageVisibility{
			Enabled:          true,
			StatusPageIDs:    []string{page.ID},
			ComponentConfigs: configs,
		},
	}
	created, err := client.CreateMaintenanceWindow(ctx, pending)
	if errors.Is(err, checkly.ErrForbidden) || errors.Is(err, checkly.ErrPlanLimit) {
		t.Skipf("the account cannot show maintenance windows on status pages: %v", err)
	}
	if err != nil {
		t.Fatalf("failed to create maintenance window: %v", err)
	}
	defer func() {
		_ = client.DeleteMaintenanceWindow(ctx, created.ID)
	}()

	read, err := client.GetMaintenanceWindow(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get maintenance window: %v", err)
	}
	if !cmp.Equal(configs, read.StatusPageVisibility.ComponentConfigs) {
		t.Fatalf("component configs after create: %s", cmp.Diff(configs, read.StatusPageVisibility.ComponentConfigs))
	}

	pending = *read
	pending.Name = "Bar maintenance"
	updated, err := client.UpdateMaintenanceWindow(ctx, created.ID, pending)
	if err != nil {
		t.Fatalf("failed to update maintenance window with the component configs read: %v", err)
	}
	if !cmp.Equal(configs, updated.StatusPageVisibility.ComponentConfigs) {
		t.Fatalf("component configs after a round trip: %s", cmp.Diff(configs, updated.StatusPageVisibility.ComponentConfigs))
	}

	pending.StatusPageVisibility.StatusPageIDs = []string{}
	pending.StatusPageVisibility.ComponentConfigs = []checkly.MaintenanceWindowComponentConfig{}
	updated, err = client.UpdateMaintenanceWindow(ctx, created.ID, pending)
	if err != nil {
		t.Fatalf("failed to update maintenance window with empty component configs: %v", err)
	}
	if len(updated.StatusPageVisibility.ComponentConfigs) != 0 || len(updated.StatusPageVisibility.StatusPageIDs) != 0 {
		t.Fatalf("expected the status page unlinked, got %+v", updated.StatusPageVisibility)
	}
}

// createResetTestGroup creates a group for the update-reset tests, so that
// their monitors can carry a group order, and deletes it when the test ends.
func createResetTestGroup(ctx context.Context, t *testing.T, client checkly.Client) *checkly.GroupV2 {
	t.Helper()
	group, err := client.CreateGroupV2(ctx, checkly.GroupV2{
		Name:        "Update reset test group",
		Concurrency: 3,
		Tags:        []string{},
		AlertSettings: &checkly.AlertSettings{
			EscalationType: checkly.EscalationTypeRunBased,
			RunBasedEscalation: checkly.RunBasedEscalation{
				FailedRunThreshold: 1,
			},
		},
		Locations: []string{"us-east-1"},
	})
	if err != nil {
		t.Fatalf("failed to create group: %v", err)
	}
	t.Cleanup(func() {
		_ = client.DeleteGroupV2(ctx, group.ID)
	})
	return group
}

// TestTracerouteMonitorUpdateResetsFields verifies that an update with unset
// request fields replaces the stored values rather than keeping them. The
// fields are sent as null, so a read returns either the zero value or the
// API default, never the value set on create.
func TestTracerouteMonitorUpdateResetsFields(t *testing.T) {
	ctx := context.TODO()
	client := setupClient(t)

	ptrLookup := true
	pending := checkly.TracerouteMonitor{
		Name:      "Update reset traceroute monitor",
		Frequency: 10,
		Locations: []string{"us-east-1"},
		Tags:      []string{"reset"},
		Request: checkly.TracerouteRequest{
			URL:            "api.checklyhq.com",
			Protocol:       "UDP",
			Port:           33435,
			MaxHops:        20,
			MaxUnknownHops: 10,
			PtrLookup:      &ptrLookup,
			Timeout:        5,
			Assertions: []checkly.Assertion{{
				Source:     "RESPONSE_TIME",
				Property:   "avg",
				Comparison: "LESS_THAN",
				Target:     "1000",
			}},
		},
	}
	created, err := client.CreateTracerouteMonitor(ctx, pending)
	if err != nil {
		t.Fatalf("failed to create traceroute monitor: %v", err)
	}
	defer func() {
		_ = client.DeleteTracerouteMonitor(ctx, created.ID)
	}()

	update := pending
	update.Tags = nil
	update.Request = checkly.TracerouteRequest{URL: pending.Request.URL, PtrLookup: &ptrLookup}
	if _, err := client.UpdateTracerouteMonitor(ctx, created.ID, update); err != nil {
		t.Fatalf("failed to update traceroute monitor: %v", err)
	}

	got, err := client.GetTracerouteMonitor(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get traceroute monitor: %v", err)
	}
	request := got.Request
	if request.Protocol != "" && request.Protocol != "TCP" {
		t.Errorf("protocol = %q, want empty or TCP", request.Protocol)
	}
	if request.Port != 0 && request.Port != 443 {
		t.Errorf("port = %d, want 0 or 443", request.Port)
	}
	if request.MaxHops != 0 && request.MaxHops != 30 {
		t.Errorf("maxHops = %d, want 0 or 30", request.MaxHops)
	}
	if request.MaxUnknownHops != 0 && request.MaxUnknownHops != 15 {
		t.Errorf("maxUnknownHops = %d, want 0 or 15", request.MaxUnknownHops)
	}
	if request.Timeout != 0 && request.Timeout != 10 {
		t.Errorf("timeout = %d, want 0 or 10", request.Timeout)
	}
	if len(request.Assertions) != 0 {
		t.Errorf("assertions = %v, want none", request.Assertions)
	}
	if len(got.Tags) != 0 {
		t.Errorf("tags = %v, want none", got.Tags)
	}
}

// TestGRPCMonitorUpdateResetsFields verifies that an update with unset fields
// replaces the stored timeout, tags and group order, and that an update with
// an empty mode applies its BEHAVIOR-mode fields.
func TestGRPCMonitorUpdateResetsFields(t *testing.T) {
	ctx := context.TODO()
	client := setupClient(t)
	group := createResetTestGroup(ctx, t, client)

	pending := checkly.GRPCMonitor{
		Name:       "Update reset gRPC monitor",
		Frequency:  10,
		Locations:  []string{"us-east-1"},
		Tags:       []string{"reset"},
		GroupID:    group.ID,
		GroupOrder: 2,
		Request: checkly.GRPCRequest{
			URL:     "grpc.example.com",
			Port:    443,
			Timeout: 30,
			GRPCConfig: checkly.GRPCConfig{
				Mode:   "BEHAVIOR",
				Method: "example.Service/Old",
			},
		},
	}
	created, err := client.CreateGRPCMonitor(ctx, pending)
	if err != nil {
		t.Fatalf("failed to create gRPC monitor: %v", err)
	}
	defer func() {
		_ = client.DeleteGRPCMonitor(ctx, created.ID)
	}()

	update := pending
	update.Tags = nil
	update.GroupOrder = 0
	update.Request.Timeout = 0
	update.Request.GRPCConfig = checkly.GRPCConfig{Method: "example.Service/New"}
	if _, err := client.UpdateGRPCMonitor(ctx, created.ID, update); err != nil {
		t.Fatalf("failed to update gRPC monitor: %v", err)
	}

	got, err := client.GetGRPCMonitor(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get gRPC monitor: %v", err)
	}
	if got.Request.Timeout != 0 && got.Request.Timeout != 60 {
		t.Errorf("timeout = %d, want 0 or 60", got.Request.Timeout)
	}
	if len(got.Tags) != 0 {
		t.Errorf("tags = %v, want none", got.Tags)
	}
	if got.GroupOrder != 0 {
		t.Errorf("groupOrder = %d, want 0", got.GroupOrder)
	}
	if got.Request.GRPCConfig.Mode != "BEHAVIOR" {
		t.Errorf("mode = %q, want BEHAVIOR", got.Request.GRPCConfig.Mode)
	}
	if got.Request.GRPCConfig.Method != "example.Service/New" {
		t.Errorf("method = %q, want the updated method", got.Request.GRPCConfig.Method)
	}
}

// TestSSLMonitorUpdateResetsFields verifies that an update with unset tags
// and group order replaces the stored values.
func TestSSLMonitorUpdateResetsFields(t *testing.T) {
	ctx := context.TODO()
	client := setupClient(t)
	group := createResetTestGroup(ctx, t, client)

	pending := checkly.SSLMonitor{
		Name:       "Update reset SSL monitor",
		Frequency:  10,
		Locations:  []string{"us-east-1"},
		Tags:       []string{"reset"},
		GroupID:    group.ID,
		GroupOrder: 2,
		Request: checkly.SSLRequest{
			SSLConfig: checkly.SSLConfig{Hostname: "api.checklyhq.com"},
		},
	}
	created, err := client.CreateSSLMonitor(ctx, pending)
	if err != nil {
		t.Fatalf("failed to create SSL monitor: %v", err)
	}
	defer func() {
		_ = client.DeleteSSLMonitor(ctx, created.ID)
	}()

	update := pending
	update.Tags = nil
	update.GroupOrder = 0
	if _, err := client.UpdateSSLMonitor(ctx, created.ID, update); err != nil {
		t.Fatalf("failed to update SSL monitor: %v", err)
	}

	got, err := client.GetSSLMonitor(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get SSL monitor: %v", err)
	}
	if len(got.Tags) != 0 {
		t.Errorf("tags = %v, want none", got.Tags)
	}
	if got.GroupOrder != 0 {
		t.Errorf("groupOrder = %d, want 0", got.GroupOrder)
	}
}

// TestStatusPageUpdateResetsFields verifies that an update with an empty
// custom domain and logo removes the stored values.
func TestStatusPageUpdateResetsFields(t *testing.T) {
	ctx := context.TODO()
	client := setupClient(t)

	suffix := time.Now().UnixNano()
	pending := checkly.StatusPage{
		Name:         "Update reset status page",
		URL:          fmt.Sprintf("sdk-update-reset-%d", suffix),
		CustomDomain: fmt.Sprintf("sdk-update-reset-%d.checklyhq.com", suffix),
		Logo:         "https://www.checklyhq.com/images/logo.png",
		Cards:        []checkly.StatusPageCard{},
	}
	created, err := client.CreateStatusPage(ctx, pending)
	if err != nil {
		t.Fatalf("failed to create status page: %v", err)
	}
	defer func() {
		_ = client.DeleteStatusPage(ctx, created.ID)
	}()

	update := pending
	update.CustomDomain = ""
	update.Logo = ""
	if _, err := client.UpdateStatusPage(ctx, created.ID, update); err != nil {
		t.Fatalf("failed to update status page: %v", err)
	}

	got, err := client.GetStatusPage(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get status page: %v", err)
	}
	if got.CustomDomain != "" {
		t.Errorf("customDomain = %q, want empty", got.CustomDomain)
	}
	if got.Logo != "" {
		t.Errorf("logo = %q, want empty", got.Logo)
	}
}

// TestCheckUpdateResetsScripts verifies that UpdateAPICheck with unset snippet
// IDs and local scripts removes the stored values.
func TestCheckUpdateResetsScripts(t *testing.T) {
	ctx := context.TODO()
	client := setupClient(t)

	snippet, err := client.CreateSnippet(ctx, checkly.Snippet{
		Name:   "Update reset snippet",
		Script: "console.log('setup')",
	})
	if err != nil {
		t.Fatalf("failed to create snippet: %v", err)
	}
	defer func() {
		_ = client.DeleteSnippet(ctx, snippet.ID)
	}()

	localSetup := "console.log('local setup')"
	localTearDown := "console.log('local teardown')"
	pending := checkly.APICheck{
		Name:                 "Update reset check",
		Frequency:            10,
		Locations:            []string{"us-east-1"},
		DegradedResponseTime: 5000,
		MaxResponseTime:      20000,
		Request: checkly.Request{
			Method: "GET",
			URL:    "https://api.checklyhq.com",
		},
		SetupSnippetID:      snippet.ID,
		TearDownSnippetID:   snippet.ID,
		LocalSetupScript:    &localSetup,
		LocalTearDownScript: &localTearDown,
	}
	created, err := client.CreateAPICheck(ctx, pending)
	if err != nil {
		t.Fatalf("failed to create check: %v", err)
	}
	defer func() {
		_ = client.DeleteAPICheck(ctx, created.ID)
	}()

	update := pending
	update.SetupSnippetID = 0
	update.TearDownSnippetID = 0
	update.LocalSetupScript = nil
	update.LocalTearDownScript = nil
	if _, err := client.UpdateAPICheck(ctx, created.ID, update); err != nil {
		t.Fatalf("failed to update check: %v", err)
	}

	got, err := client.GetAPICheck(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get check: %v", err)
	}
	if got.SetupSnippetID != 0 || got.TearDownSnippetID != 0 {
		t.Errorf("snippet IDs = %d, %d, want none", got.SetupSnippetID, got.TearDownSnippetID)
	}
	if got.LocalSetupScript != nil {
		t.Errorf("local setup script = %q, want none", *got.LocalSetupScript)
	}
	if got.LocalTearDownScript != nil {
		t.Errorf("local teardown script = %q, want none", *got.LocalTearDownScript)
	}
}

// TestUnsetListsAreAccepted verifies that create and update succeed when the
// lists the API rejects as null are left nil: the client sends them as [].
func TestUnsetListsAreAccepted(t *testing.T) {
	ctx := context.TODO()
	client := setupClient(t)

	alertSettings := checkly.AlertSettings{
		EscalationType: checkly.EscalationTypeRunBased,
		RunBasedEscalation: checkly.RunBasedEscalation{
			FailedRunThreshold: 1,
		},
	}

	t.Run("APICheck", func(t *testing.T) {
		check := checkly.APICheck{
			Name:                 "Unset lists API check",
			Frequency:            10,
			Locations:            []string{"us-east-1"},
			DegradedResponseTime: 5000,
			MaxResponseTime:      20000,
			AlertSettings:        &alertSettings,
			Request: checkly.Request{
				Method: "GET",
				URL:    "https://api.checklyhq.com",
			},
		}
		created, err := client.CreateAPICheck(ctx, check)
		if err != nil {
			t.Fatalf("failed to create API check: %v", err)
		}
		defer func() {
			_ = client.DeleteAPICheck(ctx, created.ID)
		}()
		if _, err := client.UpdateAPICheck(ctx, created.ID, check); err != nil {
			t.Fatalf("failed to update API check: %v", err)
		}
	})

	t.Run("URLMonitor", func(t *testing.T) {
		monitor := checkly.URLMonitor{
			Name:      "Unset lists URL monitor",
			Frequency: 10,
			Locations: []string{"us-east-1"},
			Request: checkly.URLRequest{
				URL: "https://api.checklyhq.com",
			},
		}
		created, err := client.CreateURLMonitor(ctx, monitor)
		if err != nil {
			t.Fatalf("failed to create URL monitor: %v", err)
		}
		defer func() {
			_ = client.DeleteURLMonitor(ctx, created.ID)
		}()
		if _, err := client.UpdateURLMonitor(ctx, created.ID, monitor); err != nil {
			t.Fatalf("failed to update URL monitor: %v", err)
		}
	})

	t.Run("DNSMonitor", func(t *testing.T) {
		monitor := checkly.DNSMonitor{
			Name:                 "Unset lists DNS monitor",
			Frequency:            10,
			Locations:            []string{"us-east-1"},
			DegradedResponseTime: 500,
			MaxResponseTime:      1000,
			Request: checkly.DNSRequest{
				RecordType: "A",
				Query:      "welcome.checklyhq.com",
			},
		}
		created, err := client.CreateDNSMonitor(ctx, monitor)
		if err != nil {
			t.Fatalf("failed to create DNS monitor: %v", err)
		}
		defer func() {
			_ = client.DeleteDNSMonitor(ctx, created.ID)
		}()
		if _, err := client.UpdateDNSMonitor(ctx, created.ID, monitor); err != nil {
			t.Fatalf("failed to update DNS monitor: %v", err)
		}
	})

	t.Run("ICMPMonitor", func(t *testing.T) {
		monitor := checkly.ICMPMonitor{
			Name:                        "Unset lists ICMP monitor",
			Frequency:                   10,
			Locations:                   []string{"us-east-1"},
			DegradedPacketLossThreshold: 10,
			MaxPacketLossThreshold:      20,
			Request: checkly.ICMPRequest{
				Hostname: "welcome.checklyhq.com",
			},
		}
		created, err := client.CreateICMPMonitor(ctx, monitor)
		if err != nil {
			t.Fatalf("failed to create ICMP monitor: %v", err)
		}
		defer func() {
			_ = client.DeleteICMPMonitor(ctx, created.ID)
		}()
		if _, err := client.UpdateICMPMonitor(ctx, created.ID, monitor); err != nil {
			t.Fatalf("failed to update ICMP monitor: %v", err)
		}
	})

	t.Run("GroupV2", func(t *testing.T) {
		group := checkly.GroupV2{
			Name:          "Unset lists group v2",
			Concurrency:   3,
			Locations:     []string{"us-east-1"},
			AlertSettings: &alertSettings,
		}
		created, err := client.CreateGroupV2(ctx, group)
		if err != nil {
			t.Fatalf("failed to create group v2: %v", err)
		}
		defer func() {
			_ = client.DeleteGroupV2(ctx, created.ID)
		}()
		if _, err := client.UpdateGroupV2(ctx, created.ID, group); err != nil {
			t.Fatalf("failed to update group v2: %v", err)
		}
	})

	t.Run("StatusPage", func(t *testing.T) {
		page := checkly.StatusPage{
			Name:  "Unset lists status page",
			URL:   fmt.Sprintf("sdk-unset-lists-%d", time.Now().UnixNano()),
			Cards: []checkly.StatusPageCard{{Name: "No services"}},
		}
		created, err := client.CreateStatusPage(ctx, page)
		if err != nil {
			t.Fatalf("failed to create status page: %v", err)
		}
		defer func() {
			_ = client.DeleteStatusPage(ctx, created.ID)
		}()
		if _, err := client.UpdateStatusPage(ctx, created.ID, page); err != nil {
			t.Fatalf("failed to update status page: %v", err)
		}
	})
}

// TestMonitorPrivateLocations verifies that DNS, ICMP and traceroute monitors
// can be created with a private location, that reads return it, and that an
// update with unset private locations removes it. The private location is
// created by the test, so the account needs none, and is deleted after the
// monitors that use it.
func TestMonitorPrivateLocations(t *testing.T) {
	ctx := context.TODO()
	client := setupClient(t)

	suffix := time.Now().UnixNano()
	location, err := client.CreatePrivateLocation(ctx, checkly.PrivateLocation{
		Name: fmt.Sprintf("SDK monitor private location %d", suffix),
		// Slug names are limited to 30 characters.
		SlugName: fmt.Sprintf("sdk-mpl-%d", suffix),
		Icon:     "location",
	})
	if err != nil {
		t.Fatalf("failed to create private location: %v", err)
	}
	defer func() {
		if err := client.DeletePrivateLocation(ctx, location.ID); err != nil {
			t.Errorf("failed to delete private location %s: %v", location.ID, err)
		}
	}()
	want := []string{location.SlugName}

	assertPrivateLocations := func(t *testing.T, step string, got *[]string, want []string) {
		t.Helper()
		var values []string
		if got != nil {
			values = *got
		}
		if !cmp.Equal(want, values, cmpopts.EquateEmpty()) {
			t.Errorf("%s: privateLocations = %v, want %v", step, values, want)
		}
	}

	t.Run("DNS", func(t *testing.T) {
		pending := checkly.DNSMonitor{
			Name:             "Private location DNS monitor",
			Frequency:        10,
			Locations:        []string{"us-east-1"},
			PrivateLocations: &want,
			Request: checkly.DNSRequest{
				RecordType: "A",
				Query:      "welcome.checklyhq.com",
			},
		}
		created, err := client.CreateDNSMonitor(ctx, pending)
		if err != nil {
			t.Fatalf("failed to create DNS monitor: %v", err)
		}
		defer func() { _ = client.DeleteDNSMonitor(ctx, created.ID) }()
		assertPrivateLocations(t, "create", created.PrivateLocations, want)

		got, err := client.GetDNSMonitor(ctx, created.ID)
		if err != nil {
			t.Fatalf("failed to get DNS monitor: %v", err)
		}
		assertPrivateLocations(t, "get", got.PrivateLocations, want)

		update := pending
		update.PrivateLocations = nil
		if _, err := client.UpdateDNSMonitor(ctx, created.ID, update); err != nil {
			t.Fatalf("failed to update DNS monitor: %v", err)
		}
		got, err = client.GetDNSMonitor(ctx, created.ID)
		if err != nil {
			t.Fatalf("failed to get DNS monitor: %v", err)
		}
		assertPrivateLocations(t, "get after update", got.PrivateLocations, nil)
	})

	t.Run("ICMP", func(t *testing.T) {
		pending := checkly.ICMPMonitor{
			Name:                        "Private location ICMP monitor",
			Frequency:                   10,
			Locations:                   []string{"us-east-1"},
			PrivateLocations:            &want,
			DegradedPacketLossThreshold: 10,
			MaxPacketLossThreshold:      20,
			Request: checkly.ICMPRequest{
				Hostname: "welcome.checklyhq.com",
			},
		}
		created, err := client.CreateICMPMonitor(ctx, pending)
		if err != nil {
			t.Fatalf("failed to create ICMP monitor: %v", err)
		}
		defer func() { _ = client.DeleteICMPMonitor(ctx, created.ID) }()
		assertPrivateLocations(t, "create", created.PrivateLocations, want)

		got, err := client.GetICMPMonitor(ctx, created.ID)
		if err != nil {
			t.Fatalf("failed to get ICMP monitor: %v", err)
		}
		assertPrivateLocations(t, "get", got.PrivateLocations, want)

		update := pending
		update.PrivateLocations = nil
		if _, err := client.UpdateICMPMonitor(ctx, created.ID, update); err != nil {
			t.Fatalf("failed to update ICMP monitor: %v", err)
		}
		got, err = client.GetICMPMonitor(ctx, created.ID)
		if err != nil {
			t.Fatalf("failed to get ICMP monitor: %v", err)
		}
		assertPrivateLocations(t, "get after update", got.PrivateLocations, nil)
	})

	t.Run("Traceroute", func(t *testing.T) {
		pending := checkly.TracerouteMonitor{
			Name:             "Private location traceroute monitor",
			Frequency:        10,
			Locations:        []string{"us-east-1"},
			PrivateLocations: &want,
			Request: checkly.TracerouteRequest{
				URL: "api.checklyhq.com",
			},
		}
		created, err := client.CreateTracerouteMonitor(ctx, pending)
		if err != nil {
			t.Fatalf("failed to create traceroute monitor: %v", err)
		}
		defer func() { _ = client.DeleteTracerouteMonitor(ctx, created.ID) }()
		assertPrivateLocations(t, "create", created.PrivateLocations, want)

		got, err := client.GetTracerouteMonitor(ctx, created.ID)
		if err != nil {
			t.Fatalf("failed to get traceroute monitor: %v", err)
		}
		assertPrivateLocations(t, "get", got.PrivateLocations, want)

		update := pending
		update.PrivateLocations = nil
		if _, err := client.UpdateTracerouteMonitor(ctx, created.ID, update); err != nil {
			t.Fatalf("failed to update traceroute monitor: %v", err)
		}
		got, err = client.GetTracerouteMonitor(ctx, created.ID)
		if err != nil {
			t.Fatalf("failed to get traceroute monitor: %v", err)
		}
		assertPrivateLocations(t, "get after update", got.PrivateLocations, nil)
	})
}

// TestBadRequestErrorsIntegration verifies that the API's two kinds of 400
// response are told apart: a schema failure matches ErrValidation, and a
// rule failure in a handler matches only ErrBadRequest.
func TestBadRequestErrorsIntegration(t *testing.T) {
	ctx := context.TODO()
	client := setupClient(t)

	check := checkly.APICheck{
		Name:                 "Bad request errors integration check",
		Frequency:            7,
		Locations:            []string{"us-east-1"},
		Tags:                 []string{"integration"},
		DegradedResponseTime: 5000,
		MaxResponseTime:      20000,
		Request: checkly.Request{
			Method: "GET",
			URL:    "https://api.checklyhq.com",
		},
	}
	_, err := client.CreateAPICheck(ctx, check)
	if !errors.Is(err, checkly.ErrValidation) || !errors.Is(err, checkly.ErrBadRequest) {
		t.Fatalf("want ErrValidation and ErrBadRequest for an invalid frequency, got %v", err)
	}
	var apiErr *checkly.APIError
	if !errors.As(err, &apiErr) || len(apiErr.Details) == 0 {
		t.Errorf("want Details to name the field, got %v", err)
	}

	check.Frequency = 10
	created, err := client.CreateAPICheck(ctx, check)
	if err != nil {
		t.Fatalf("failed to create check: %v", err)
	}
	t.Cleanup(func() {
		_ = client.DeleteAPICheck(context.Background(), created.ID)
	})

	_, err = client.GetCheckResults(ctx, created.ID, &checkly.CheckResultsFilter{From: 2000, To: 1000})
	if !errors.Is(err, checkly.ErrBadRequest) {
		t.Fatalf("want ErrBadRequest for a range that ends before it starts, got %v", err)
	}
	if errors.Is(err, checkly.ErrValidation) {
		t.Errorf("want a rule failure not to match ErrValidation, got %v", err)
	}
}

// TestGetCheckResultsIntegration verifies that the API accepts the requests
// GetCheckResults sends, including the HasFailures filter, and that an empty
// page decodes. The check is never activated, so it has no results.
func TestGetCheckResultsIntegration(t *testing.T) {
	ctx := context.TODO()
	client := setupClient(t)

	created, err := client.CreateAPICheck(ctx, checkly.APICheck{
		Name:                 "GetCheckResults integration check",
		Frequency:            10,
		Locations:            []string{"us-east-1"},
		Tags:                 []string{"integration"},
		DegradedResponseTime: 5000,
		MaxResponseTime:      20000,
		Request: checkly.Request{
			Method: "GET",
			URL:    "https://api.checklyhq.com",
		},
	})
	if err != nil {
		t.Fatalf("failed to create check: %v", err)
	}
	t.Cleanup(func() {
		_ = client.DeleteAPICheck(context.Background(), created.ID)
	})

	tests := map[string]*checkly.CheckResultsFilter{
		"nil filter":   nil,
		"has failures": {HasFailures: true},
	}
	for name, filter := range tests {
		t.Run(name, func(t *testing.T) {
			page, err := client.GetCheckResults(ctx, created.ID, filter)
			if err != nil {
				t.Fatalf("failed to get check results: %v", err)
			}
			if len(page.Entries) != 0 {
				t.Errorf("got %d results, want none", len(page.Entries))
			}
			if page.NextID != "" {
				t.Errorf("NextID = %q, want empty", page.NextID)
			}
		})
	}
}
