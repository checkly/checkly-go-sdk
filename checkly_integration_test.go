//go:build integration
// +build integration

package checkly_test

import (
	"context"
	"io"
	"os"
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
		t.Error("'CHECKLY_API_KEY' must be set for integration tests")
	}
	accountId := os.Getenv("CHECKLY_ACCOUNT_ID")
	if accountId == "" {
		t.Error("'CHECKLY_ACCOUNT_ID' must be set for integration tests")
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
		t.Error(err)
	}
	defer client.Delete(context.Background(), gotCheck.ID)
	if !cmp.Equal(wantCheck, *gotCheck, ignoreCheckFields) {
		t.Error(cmp.Diff(wantCheck, *gotCheck, ignoreCheckFields))
	}
}

func TestGetIntegration(t *testing.T) {
	client := setupClient(t)
	check, err := client.Create(context.Background(), wantCheck)
	if err != nil {
		t.Error(err)
	}
	defer client.Delete(context.Background(), check.ID)
	gotCheck, err := client.Get(context.Background(), check.ID)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(wantCheck, *gotCheck, ignoreCheckFields) {
		t.Error(cmp.Diff(wantCheck, *gotCheck, ignoreCheckFields))
	}
}

func TestUpdateIntegration(t *testing.T) {
	t.Parallel()
	client := setupClient(t)
	check, err := client.Create(context.Background(), wantCheck)
	if err != nil {
		t.Error(err)
	}
	defer client.Delete(context.Background(), check.ID)
	updatedCheck := wantCheck
	updatedCheck.Name = "integrationTestUpdate"
	gotCheck, err := client.Update(context.Background(), check.ID, updatedCheck)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(updatedCheck, *gotCheck, ignoreCheckFields) {
		t.Error(cmp.Diff(updatedCheck, *gotCheck, ignoreCheckFields))
	}
}

func TestDeleteIntegration(t *testing.T) {
	t.Parallel()
	client := setupClient(t)
	check, err := client.Create(context.Background(), wantCheck)
	if err != nil {
		t.Error(err)
	}
	if err := client.Delete(context.Background(), check.ID); err != nil {
		t.Error(err)
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
		t.Error(err.Error())
		return
	}
	wantGroupCopy.AlertChannelSubscriptions[0].ChannelID = ac.ID
	gotGroup, err := client.CreateGroup(context.Background(), wantGroupCopy)
	if err != nil {
		t.Error(err)
	}
	defer client.DeleteGroup(context.Background(), gotGroup.ID)
	// These are set by the APIs
	ignored := cmpopts.IgnoreFields(checkly.Group{}, "ID", "AlertChannelSubscriptions", "AlertSettings.SSLCertificates", "PrivateLocations")
	if !cmp.Equal(wantGroupCopy, *gotGroup, ignored) {
		t.Error(cmp.Diff(wantGroupCopy, *gotGroup, ignored))
	}
}

func TestGetGroupIntegration(t *testing.T) {
	client := setupClient(t)
	ac, err := makeTestAlertChannel(client)
	wantGroupCopy := wantGroup
	wantGroupCopy.AlertChannelSubscriptions[0].ChannelID = ac.ID
	group, err := client.CreateGroup(context.Background(), wantGroupCopy)
	if err != nil {
		t.Error(err)
	}
	defer client.DeleteGroup(context.Background(), group.ID)
	gotGroup, err := client.GetGroup(context.Background(), group.ID)
	if err != nil {
		t.Error(err)
	}
	// These are set by the API
	wantGroupCopy.AlertChannelSubscriptions = gotGroup.AlertChannelSubscriptions
	if !cmp.Equal(wantGroupCopy, *gotGroup, ignoreGroupFields) {
		t.Error(cmp.Diff(wantGroupCopy, *gotGroup, ignoreGroupFields))
	}
}

//Dashboard

func TestCreateDashboardIntegration(t *testing.T) {
	client := setupClient(t)

	gotDashboard, err := client.CreateDashboard(context.Background(), testDashboard)
	if err != nil {
		t.Error(err)
	}
	defer client.DeleteDashboard(context.Background(), gotDashboard.DashboardID)
	if !cmp.Equal(testDashboard, *gotDashboard, ignoreDashboardFields) {
		t.Error(cmp.Diff(testDashboard, *gotDashboard, ignoreDashboardFields))
	}
}

func TestGetDashboardIntegration(t *testing.T) {
	client := setupClient(t)
	dash, err := client.CreateDashboard(context.Background(), testDashboard)
	if err != nil {
		t.Error(err)
	}
	defer client.DeleteDashboard(context.Background(), dash.DashboardID)
	gotDashboard, err := client.GetDashboard(context.Background(), dash.DashboardID)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testDashboard, *gotDashboard, ignoreDashboardFields) {
		t.Error(cmp.Diff(testDashboard, *gotDashboard, ignoreDashboardFields))
	}
}

//Maintenance Windows

func TestCreateMaintenanceWindowIntegration(t *testing.T) {
	client := setupClient(t)

	gotMaintenanceWindow, err := client.CreateMaintenanceWindow(context.Background(), testMaintenanceWindow)
	if err != nil {
		t.Error(err)
	}
	defer client.DeleteMaintenanceWindow(context.Background(), gotMaintenanceWindow.ID)
	if !cmp.Equal(testMaintenanceWindow, *gotMaintenanceWindow, ignoreMaintenanceWindowFields) {
		t.Error(cmp.Diff(testMaintenanceWindow, *gotMaintenanceWindow, ignoreMaintenanceWindowFields))
	}
}

func TestGetMaintenanceWindowIntegration(t *testing.T) {
	client := setupClient(t)
	dash, err := client.CreateMaintenanceWindow(context.Background(), testMaintenanceWindow)
	if err != nil {
		t.Error(err)
	}
	defer client.DeleteMaintenanceWindow(context.Background(), dash.ID)
	gotMaintenanceWindow, err := client.GetMaintenanceWindow(context.Background(), dash.ID)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testMaintenanceWindow, *gotMaintenanceWindow, ignoreMaintenanceWindowFields) {
		t.Error(cmp.Diff(testMaintenanceWindow, *gotMaintenanceWindow, ignoreMaintenanceWindowFields))
	}
}

//TriggerCheck

func TestCreateTriggerCheckIntegration(t *testing.T) {
	client := setupClient(t)

	gotCheck, err := client.Create(context.Background(), wantCheck)
	if err != nil {
		t.Error(err)
	}

	gotTriggerCheck, err := client.CreateTriggerCheck(context.Background(), gotCheck.ID)
	if err != nil {
		t.Error(err)
	}
	defer client.Delete(context.Background(), gotCheck.ID)

	if !cmp.Equal(gotCheck.ID, gotTriggerCheck.CheckId, ignoreTriggerCheck) {
		t.Error(cmp.Diff(gotCheck.ID, gotTriggerCheck.CheckId, ignoreTriggerCheck))
	}
}

func TestGetTriggerCheckIntegration(t *testing.T) {
	client := setupClient(t)
	gotCheck, err := client.Create(context.Background(), wantCheck)
	if err != nil {
		t.Error(err)
	}
	tc, err := client.CreateTriggerCheck(context.Background(), gotCheck.ID)
	if err != nil {
		t.Error(err)
	}
	gotTriggerCheck, err := client.GetTriggerCheck(context.Background(), tc.CheckId)
	if err != nil {
		t.Error(err)
	}
	defer client.Delete(context.Background(), gotCheck.ID)
	if !cmp.Equal(gotCheck.ID, gotTriggerCheck.CheckId, ignoreTriggerCheck) {
		t.Error(cmp.Diff(gotCheck.ID, gotTriggerCheck.CheckId, ignoreTriggerCheck))
	}
}

//TriggerGroup

func TestCreateTriggerGroupIntegration(t *testing.T) {
	client := setupClient(t)
	gotGroup, err := client.CreateGroup(context.Background(), wantGroup)
	if err != nil {
		t.Error(err)
	}
	gotTriggerGroup, err := client.CreateTriggerGroup(context.Background(), gotGroup.ID)
	if err != nil {
		t.Error(err)
	}
	defer client.DeleteGroup(context.Background(), gotGroup.ID)
	if !cmp.Equal(gotGroup.ID, gotTriggerGroup.GroupId, ignoreTriggerGroup) {
		t.Error(cmp.Diff(gotGroup.ID, gotTriggerGroup.GroupId, ignoreTriggerGroup))
	}
}

func TestGetTriggerGroupIntegration(t *testing.T) {
	client := setupClient(t)
	gotGroup, err := client.CreateGroup(context.Background(), wantGroup)
	if err != nil {
		t.Error(err)
	}
	tc, err := client.CreateTriggerGroup(context.Background(), gotGroup.ID)
	if err != nil {
		t.Error(err)
	}
	defer client.DeleteTriggerGroup(context.Background(), tc.GroupId)
	gotTriggerGroup, err := client.GetTriggerGroup(context.Background(), tc.GroupId)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(gotGroup.ID, gotTriggerGroup.GroupId, ignoreTriggerGroup) {
		t.Error(cmp.Diff(gotGroup.ID, gotTriggerGroup.GroupId, ignoreTriggerGroup))
	}
}

// PrivateLocations

func TestCreatePrivateLocationIntegration(t *testing.T) {
	client := setupClient(t)

	gotPrivateLocation, err := client.CreatePrivateLocation(context.Background(), testPrivateLocation)
	if err != nil {
		t.Error(err)
	}
	defer client.DeletePrivateLocation(context.Background(), gotPrivateLocation.ID)
	if !cmp.Equal(testPrivateLocation, *gotPrivateLocation, ignorePrivateLocationFields) {
		t.Error(cmp.Diff(testPrivateLocation, *gotPrivateLocation, ignorePrivateLocationFields))
	}
}
func TestGetPrivateLocationIntegration(t *testing.T) {
	client := setupClient(t)
	pl, err := client.CreatePrivateLocation(context.Background(), testPrivateLocation)
	if err != nil {
		t.Error(err)
	}
	defer client.DeletePrivateLocation(context.Background(), pl.ID)
	gotPrivateLocation, err := client.GetPrivateLocation(context.Background(), pl.ID)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testPrivateLocation, *gotPrivateLocation, ignorePrivateLocationFields) {
		t.Error(cmp.Diff(testPrivateLocation, *gotPrivateLocation, ignorePrivateLocationFields))
	}
}
