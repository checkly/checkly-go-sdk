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
		Name: "Foo monitor",
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
		Name: "Bar monitor",
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
