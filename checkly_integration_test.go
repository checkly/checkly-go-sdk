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
	return checkly.NewClient(
		baseUrl,
		apiKey,
		nil,
		debug,
	)

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
	ignored := cmpopts.IgnoreFields(checkly.Group{}, "ID", "AlertChannelSubscriptions")
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

	gotTriggerCheck, err := client.CreateTriggerCheck(context.Background(), testTriggerCheck.CheckId)
	if err != nil {
		t.Error(err)
	}
	defer client.DeleteTriggerCheck(context.Background(), gotTriggerCheck.CheckId, gotTriggerCheck.Token)
	if !cmp.Equal(testTriggerCheck, *gotTriggerCheck, ignoreTriggerCheck) {
		t.Error(cmp.Diff(testTriggerCheck, *gotTriggerCheck, ignoreTriggerCheck))
	}
}

func TestGetTriggerCheckIntegration(t *testing.T) {
	client := setupClient(t)
	tc, err := client.CreateTriggerCheck(context.Background(), testTriggerCheck.CheckId)
	if err != nil {
		t.Error(err)
	}
	defer client.DeleteTriggerCheck(context.Background(), tc.CheckId, tc.Token)
	gotTriggerCheck, err := client.GetTriggerCheck(context.Background(), tc.CheckId)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testTriggerCheck, *gotTriggerCheck, ignoreTriggerCheck) {
		t.Error(cmp.Diff(testTriggerCheck, *gotTriggerCheck, ignoreTriggerCheck))
	}
}

//TriggerGroup

func TestCreateTriggerGroupIntegration(t *testing.T) {
	client := setupClient(t)

	gotTriggerGroup, err := client.CreateTriggerGroup(context.Background(), testTriggerGroup.GroupId)
	if err != nil {
		t.Error(err)
	}
	defer client.DeleteTriggerGroup(context.Background(), gotTriggerGroup.GroupId, gotTriggerGroup.Token)
	if !cmp.Equal(testTriggerGroup, *gotTriggerGroup, ignoreTriggerGroup) {
		t.Error(cmp.Diff(testTriggerGroup, *gotTriggerGroup, ignoreTriggerGroup))
	}
}

func TestGetTriggerGroupIntegration(t *testing.T) {
	client := setupClient(t)
	tc, err := client.CreateTriggerGroup(context.Background(), testTriggerGroup.GroupId)
	if err != nil {
		t.Error(err)
	}
	defer client.DeleteTriggerGroup(context.Background(), tc.GroupId, tc.Token)
	gotTriggerGroup, err := client.GetTriggerGroup(context.Background(), tc.GroupId)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(testTriggerGroup, *gotTriggerGroup, ignoreTriggerGroup) {
		t.Error(cmp.Diff(testTriggerGroup, *gotTriggerGroup, ignoreTriggerGroup))
	}
}
