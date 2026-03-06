//go:build integration
// +build integration

package checkly_test

import (
	"context"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	checkly "github.com/checkly/checkly-go-sdk"
)

func setupClient(t *testing.T) checkly.Client {
	var debug io.Writer // to enable debug output set to => os.Stdout
	baseUrl := os.Getenv("CHECKLY_API_URL")
	if baseUrl == "" {
		baseUrl = "http://localhost:3000"
	}
	apiKey := os.Getenv("CHECKLY_API_KEY")
	if apiKey == "" {
		t.Fatal("'CHECKLY_API_KEY' must be set for integration tests")
	}
	accountId := os.Getenv("CHECKLY_ACCOUNT_ID")
	if accountId == "" {
		t.Fatal("'CHECKLY_ACCOUNT_ID' must be set for integration tests")
	}

	client := checkly.NewClient(
		baseUrl,
		apiKey,
		nil,
		debug,
	)

	client.SetAccountId(accountId)
	return client
}

func TestCreateIntegration(t *testing.T) {
	client := setupClient(t)

	gotCheck, err := client.Create(context.Background(), wantCheck)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Delete(context.Background(), gotCheck.ID)
	if !cmp.Equal(wantCheck, *gotCheck, ignoreCheckFields) {
		t.Fatal(cmp.Diff(wantCheck, *gotCheck, ignoreCheckFields))
	}
}

func TestCheckGroupUnset(t *testing.T) {
	ctx := context.TODO()

	client := setupClient(t)

	group, err := client.CreateGroup(ctx, checkly.Group{
		Name:        "Test Group",
		Concurrency: 3,
		Tags:        []string{},
		AlertSettings: checkly.AlertSettings{
			EscalationType: checkly.RunBased,
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
		_ = client.DeleteGroup(ctx, group.ID)
	}()

	pendingCheck := checkly.Check{
		Name:      "Foo check",
		Type:      checkly.TypeAPI,
		Frequency: 10,
		Request: checkly.Request{
			Method:          "GET",
			URL:             "https://api.checklyhq.com",
			Headers:         []checkly.KeyValue{},
			QueryParameters: []checkly.KeyValue{},
			Assertions:      []checkly.Assertion{},
		},
		AlertSettings: checkly.AlertSettings{
			EscalationType: checkly.RunBased,
			RunBasedEscalation: checkly.RunBasedEscalation{
				FailedRunThreshold: 1,
			},
		},
		Locations: []string{"us-east-1"},
		GroupID:   group.ID,
	}

	createdCheck, err := client.CreateCheck(ctx, pendingCheck)
	if err != nil {
		t.Fatalf("failed to create check: %v", err)
	}

	defer func() {
		_ = client.DeleteTCPMonitor(ctx, createdCheck.ID)
	}()

	if createdCheck.GroupID != group.ID {
		t.Fatalf("wrong GroupID after creation")
	}

	updateCheck := pendingCheck
	updateCheck.GroupID = 0

	updatedCheck, err := client.UpdateCheck(ctx, createdCheck.ID, updateCheck)
	if err != nil {
		t.Fatalf("failed to update check: %v", err)
	}

	if updatedCheck.GroupID != 0 {
		t.Fatalf("wrong GroupID after update")
	}
}

func TestGetIntegration(t *testing.T) {
	client := setupClient(t)
	check, err := client.Create(context.Background(), wantCheck)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Delete(context.Background(), check.ID)
	gotCheck, err := client.Get(context.Background(), check.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(wantCheck, *gotCheck, ignoreCheckFields) {
		t.Fatal(cmp.Diff(wantCheck, *gotCheck, ignoreCheckFields))
	}
}

func TestUpdateIntegration(t *testing.T) {
	t.Parallel()
	client := setupClient(t)
	check, err := client.Create(context.Background(), wantCheck)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Delete(context.Background(), check.ID)
	updatedCheck := wantCheck
	updatedCheck.Name = "integrationTestUpdate"
	gotCheck, err := client.Update(context.Background(), check.ID, updatedCheck)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(updatedCheck, *gotCheck, ignoreCheckFields) {
		t.Fatal(cmp.Diff(updatedCheck, *gotCheck, ignoreCheckFields))
	}
}

func TestDeleteIntegration(t *testing.T) {
	t.Parallel()
	client := setupClient(t)
	check, err := client.Create(context.Background(), wantCheck)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Delete(context.Background(), check.ID); err != nil {
		t.Fatal(err)
	}
}

func makeTestAlertChannel(client checkly.Client) (*checkly.AlertChannel, error) {
	return client.CreateAlertChannel(
		context.Background(),
		checkly.AlertChannel{
			Type: checkly.AlertTypeEmail,
			Email: &checkly.AlertChannelEmail{
				Address: "test@example.com",
			},
		},
	)
}

func TestCreateGroupIntegration(t *testing.T) {
	client := setupClient(t)
	wantGroupCopy := wantGroup
	ac, err := makeTestAlertChannel(client)
	if err != nil {
		t.Fatal(err.Error())
	}
	wantGroupCopy.AlertChannelSubscriptions[0].ChannelID = ac.ID
	gotGroup, err := client.CreateGroup(context.Background(), wantGroupCopy)
	if err != nil {
		t.Fatal(err)
	}
	defer client.DeleteGroup(context.Background(), gotGroup.ID)
	// These are set by the APIs
	ignored := cmpopts.IgnoreFields(checkly.Group{}, "ID", "AlertChannelSubscriptions", "AlertSettings.SSLCertificates", "PrivateLocations")
	if !cmp.Equal(wantGroupCopy, *gotGroup, ignored) {
		t.Fatal(cmp.Diff(wantGroupCopy, *gotGroup, ignored))
	}
}

func TestGetGroupIntegration(t *testing.T) {
	client := setupClient(t)
	ac, err := makeTestAlertChannel(client)
	wantGroupCopy := wantGroup
	wantGroupCopy.AlertChannelSubscriptions[0].ChannelID = ac.ID
	group, err := client.CreateGroup(context.Background(), wantGroupCopy)
	if err != nil {
		t.Fatal(err)
	}
	defer client.DeleteGroup(context.Background(), group.ID)
	gotGroup, err := client.GetGroup(context.Background(), group.ID)
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
	defer client.DeleteDashboard(context.Background(), gotDashboard.DashboardID)
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
	defer client.DeleteDashboard(context.Background(), dash.DashboardID)
	gotDashboard, err := client.GetDashboard(context.Background(), dash.DashboardID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testDashboard, *gotDashboard, ignoreDashboardFields) {
		t.Fatal(cmp.Diff(testDashboard, *gotDashboard, ignoreDashboardFields))
	}
}

//Maintenance Windows

func TestCreateMaintenanceWindowIntegration(t *testing.T) {
	client := setupClient(t)

	gotMaintenanceWindow, err := client.CreateMaintenanceWindow(context.Background(), testMaintenanceWindow)
	if err != nil {
		t.Fatal(err)
	}
	defer client.DeleteMaintenanceWindow(context.Background(), gotMaintenanceWindow.ID)
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
	defer client.DeleteMaintenanceWindow(context.Background(), dash.ID)
	gotMaintenanceWindow, err := client.GetMaintenanceWindow(context.Background(), dash.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testMaintenanceWindow, *gotMaintenanceWindow, ignoreMaintenanceWindowFields) {
		t.Fatal(cmp.Diff(testMaintenanceWindow, *gotMaintenanceWindow, ignoreMaintenanceWindowFields))
	}
}

//TriggerCheck

func TestCreateTriggerCheckIntegration(t *testing.T) {
	client := setupClient(t)

	gotCheck, err := client.Create(context.Background(), wantCheck)
	if err != nil {
		t.Fatal(err)
	}

	gotTriggerCheck, err := client.CreateTriggerCheck(context.Background(), gotCheck.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Delete(context.Background(), gotCheck.ID)

	if !cmp.Equal(gotCheck.ID, gotTriggerCheck.CheckId, ignoreTriggerCheck) {
		t.Fatal(cmp.Diff(gotCheck.ID, gotTriggerCheck.CheckId, ignoreTriggerCheck))
	}
}

func TestGetTriggerCheckIntegration(t *testing.T) {
	client := setupClient(t)
	gotCheck, err := client.Create(context.Background(), wantCheck)
	if err != nil {
		t.Fatal(err)
	}
	tc, err := client.CreateTriggerCheck(context.Background(), gotCheck.ID)
	if err != nil {
		t.Fatal(err)
	}
	gotTriggerCheck, err := client.GetTriggerCheck(context.Background(), tc.CheckId)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Delete(context.Background(), gotCheck.ID)
	if !cmp.Equal(gotCheck.ID, gotTriggerCheck.CheckId, ignoreTriggerCheck) {
		t.Fatal(cmp.Diff(gotCheck.ID, gotTriggerCheck.CheckId, ignoreTriggerCheck))
	}
}

//TriggerGroup

func TestCreateTriggerGroupIntegration(t *testing.T) {
	client := setupClient(t)
	gotGroup, err := client.CreateGroup(context.Background(), wantGroup)
	if err != nil {
		t.Fatal(err)
	}
	gotTriggerGroup, err := client.CreateTriggerGroup(context.Background(), gotGroup.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer client.DeleteGroup(context.Background(), gotGroup.ID)
	if !cmp.Equal(gotGroup.ID, gotTriggerGroup.GroupId, ignoreTriggerGroup) {
		t.Fatal(cmp.Diff(gotGroup.ID, gotTriggerGroup.GroupId, ignoreTriggerGroup))
	}
}

func TestGetTriggerGroupIntegration(t *testing.T) {
	client := setupClient(t)
	gotGroup, err := client.CreateGroup(context.Background(), wantGroup)
	if err != nil {
		t.Fatal(err)
	}
	tc, err := client.CreateTriggerGroup(context.Background(), gotGroup.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer client.DeleteTriggerGroup(context.Background(), tc.GroupId)
	gotTriggerGroup, err := client.GetTriggerGroup(context.Background(), tc.GroupId)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(gotGroup.ID, gotTriggerGroup.GroupId, ignoreTriggerGroup) {
		t.Fatal(cmp.Diff(gotGroup.ID, gotTriggerGroup.GroupId, ignoreTriggerGroup))
	}
}

// PrivateLocations

func TestCreatePrivateLocationIntegration(t *testing.T) {
	client := setupClient(t)

	gotPrivateLocation, err := client.CreatePrivateLocation(context.Background(), testPrivateLocation)
	if err != nil {
		t.Fatal(err)
	}
	defer client.DeletePrivateLocation(context.Background(), gotPrivateLocation.ID)
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
	defer client.DeletePrivateLocation(context.Background(), pl.ID)
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

	pendingCheck := checkly.TCPCheck{
		Name:      "TestTCPCheckCRUD",
		Muted:     false,
		Frequency: 1,
		Locations: []string{"eu-west-1"},
		Request: checkly.TCPRequest{
			Hostname: "api.checklyhq.com",
			Port:     443,
		},
	}

	createdCheck, err := client.CreateTCPCheck(ctx, pendingCheck)
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

	_, err = client.GetTCPCheck(ctx, createdCheck.ID)
	if err != nil {
		t.Fatalf("failed to get TCP check: %v", err)
	}

	updateCheck := *createdCheck
	updateCheck.Muted = true

	updatedCheck, err := client.UpdateTCPCheck(ctx, createdCheck.ID, updateCheck)
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

	group, err := client.CreateGroup(ctx, checkly.Group{
		Name:        "Test Group",
		Concurrency: 3,
		Tags:        []string{},
		AlertSettings: checkly.AlertSettings{
			EscalationType: checkly.RunBased,
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
		_ = client.DeleteGroup(ctx, group.ID)
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

	group, err := client.CreateGroup(ctx, checkly.Group{
		Name:        "Test Group",
		Concurrency: 3,
		Tags:        []string{},
		AlertSettings: checkly.AlertSettings{
			EscalationType: checkly.RunBased,
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
		_ = client.DeleteGroup(ctx, group.ID)
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
		Name:      "Foo ICMP monitor",
		Frequency: 1,
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
		Name:      "Bar ICMP monitor",
		Frequency: 1,
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

	group, err := client.CreateGroup(ctx, checkly.Group{
		Name:        "Test Group",
		Concurrency: 3,
		Tags:        []string{},
		AlertSettings: checkly.AlertSettings{
			EscalationType: checkly.RunBased,
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
		_ = client.DeleteGroup(ctx, group.ID)
	}()

	pendingMonitor := checkly.ICMPMonitor{
		Name:      "Foo ICMP monitor",
		Frequency: 1,
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

	group, err := client.CreateGroup(ctx, checkly.Group{
		Name:        "Test Group",
		Concurrency: 3,
		Tags:        []string{},
		AlertSettings: checkly.AlertSettings{
			EscalationType: checkly.RunBased,
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
		_ = client.DeleteGroup(ctx, group.ID)
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

func TestTracerouteMonitorCRUD(t *testing.T) {
	client := setupClient(t)
	ctx := context.Background()

	ptrTrue := true
	pendingMonitor := checkly.TracerouteMonitor{
		Name:                 "Integration Test Traceroute",
		Activated:            true,
		Frequency:            60,
		DegradedResponseTime: 15000,
		MaxResponseTime:      30000,
		Locations:            []string{"us-east-1"},
		Request: checkly.TracerouteRequest{
			Hostname:       "example.com",
			Port:           443,
			IPFamily:       "IPv4",
			MaxHops:        30,
			MaxUnknownHops: 15,
			PtrLookup:      &ptrTrue,
			Timeout:        10,
			Assertions: []checkly.Assertion{
				{
					Source:     "RESPONSE_TIME",
					Property:  "avg",
					Comparison: "LESS_THAN",
					Target:    "200",
				},
			},
		},
	}

	// Create
	createdMonitor, err := client.CreateTracerouteMonitor(ctx, pendingMonitor)
	if err != nil {
		t.Fatalf("failed to create traceroute monitor: %v", err)
	}
	if createdMonitor.ID == "" {
		t.Fatal("expected non-empty ID after creation")
	}

	defer func() {
		_ = client.DeleteTracerouteMonitor(ctx, createdMonitor.ID)
	}()

	// Get
	gotMonitor, err := client.GetTracerouteMonitor(ctx, createdMonitor.ID)
	if err != nil {
		t.Fatalf("failed to get traceroute monitor: %v", err)
	}
	if gotMonitor.Name != pendingMonitor.Name {
		t.Fatalf("expected name %q, got %q", pendingMonitor.Name, gotMonitor.Name)
	}
	if gotMonitor.Request.Hostname != "example.com" {
		t.Fatalf("expected hostname 'example.com', got %q", gotMonitor.Request.Hostname)
	}
	if gotMonitor.Request.Port != 443 {
		t.Fatalf("expected port 443, got %d", gotMonitor.Request.Port)
	}

	// Update
	updateMonitor := pendingMonitor
	updateMonitor.Name = "Updated Traceroute Monitor"
	updateMonitor.Request.MaxHops = 20

	updatedTraceroute, err := client.UpdateTracerouteMonitor(ctx, createdMonitor.ID, updateMonitor)
	if err != nil {
		t.Fatalf("failed to update traceroute monitor: %v", err)
	}
	if updatedTraceroute.Name != "Updated Traceroute Monitor" {
		t.Fatalf("expected updated name, got %q", updatedTraceroute.Name)
	}
	if updatedTraceroute.Request.MaxHops != 20 {
		t.Fatalf("expected maxHops 20, got %d", updatedTraceroute.Request.MaxHops)
	}

	// Delete
	err = client.DeleteTracerouteMonitor(ctx, createdMonitor.ID)
	if err != nil {
		t.Fatalf("failed to delete traceroute monitor: %v", err)
	}
}
