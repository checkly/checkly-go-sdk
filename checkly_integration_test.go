// +build integration

package checkly_test

import (
	"os"
	"testing"

	checkly "github.com/checkly/checkly-go-sdk"
	"github.com/google/go-cmp/cmp"
)

func getAPIKey(t *testing.T) string {
	key := os.Getenv("CHECKLY_API_KEY")
	if key == "" {
		t.Fatal("'CHECKLY_API_KEY' must be set for integration tests")
	}
	return key
}

func setupClient(t *testing.T) checkly.Client {
	client := checkly.NewClient(getAPIKey(t))
	// Uncomment the following line to enable debug output
	// client.Debug = os.Stdout
	return client
}

func TestCreateIntegration(t *testing.T) {
	t.Parallel()
	client := setupClient(t)
	gotCheck, err := client.Create(wantCheck)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Delete(gotCheck.ID)
	if !cmp.Equal(wantCheck, gotCheck, ignoreCheckFields) {
		t.Error(cmp.Diff(wantCheck, gotCheck, ignoreCheckFields))
	}
}

func TestGetIntegration(t *testing.T) {
	t.Parallel()
	client := setupClient(t)
	check, err := client.Create(wantCheck)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Delete(check.ID)
	gotCheck, err := client.Get(check.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(wantCheck, gotCheck, ignoreCheckFields) {
		t.Error(cmp.Diff(wantCheck, gotCheck, ignoreCheckFields))
	}
}

func TestUpdateIntegration(t *testing.T) {
	t.Parallel()
	client := setupClient(t)
	check, err := client.Create(wantCheck)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Delete(check.ID)
	updatedCheck := wantCheck
	updatedCheck.Name = "integrationTestUpdate"
	gotCheck, err := client.Update(check.ID, updatedCheck)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(updatedCheck, gotCheck, ignoreCheckFields) {
		t.Error(cmp.Diff(updatedCheck, gotCheck, ignoreCheckFields))
	}
}

func TestDeleteIntegration(t *testing.T) {
	t.Parallel()
	client := setupClient(t)
	check, err := client.Create(wantCheck)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Delete(check.ID); err != nil {
		t.Error(err)
	}
}

func TestCreateGroupIntegration(t *testing.T) {
	t.Parallel()
	client := setupClient(t)
	wantGroupCopy := wantGroup
	gotGroup, err := client.CreateGroup(wantGroupCopy)
	if err != nil {
		t.Fatal(err)
	}
	defer client.DeleteGroup(gotGroup.ID)
	// These are set by the API
	wantGroupCopy.AlertChannelSubscriptions = gotGroup.AlertChannelSubscriptions
	if !cmp.Equal(wantGroupCopy, gotGroup, ignoreGroupFields) {
		t.Error(cmp.Diff(wantGroupCopy, gotGroup, ignoreGroupFields))
	}
}

func TestGetGroupIntegration(t *testing.T) {
	t.Parallel()
	client := setupClient(t)
	wantGroupCopy := wantGroup
	group, err := client.CreateGroup(wantGroupCopy)
	if err != nil {
		t.Fatal(err)
	}
	defer client.DeleteGroup(group.ID)
	gotGroup, err := client.GetGroup(group.ID)
	if err != nil {
		t.Error(err)
	}
	// These are set by the API
	wantGroupCopy.AlertChannelSubscriptions = gotGroup.AlertChannelSubscriptions
	if !cmp.Equal(wantGroupCopy, gotGroup, ignoreGroupFields) {
		t.Error(cmp.Diff(wantGroupCopy, gotGroup, ignoreGroupFields))
	}
}
