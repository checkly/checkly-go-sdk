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
		baseUrl = "https://localhost:3000"
	}
	apiKey := os.Getenv("CHECKLY_API_KEY")
	if apiKey == "" {
		t.Fatal("'CHECKLY_API_KEY' must be set for integration tests")
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
		t.Fatal(err)
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
		t.Fatal(err)
	}
	defer client.Delete(context.Background(), check.ID)
	gotCheck, err := client.Get(context.Background(), check.ID)
	if err != nil {
		t.Fatal(err)
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
		t.Fatal(err)
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
		t.Fatal(err)
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
		t.Fatal(err)
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
		t.Fatal(err)
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
