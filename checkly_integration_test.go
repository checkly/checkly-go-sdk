// +build integration

package checkly

import (
	// "fmt"
	"os"
	"testing"
)

func getAPIKey(t *testing.T) string {
	key := os.Getenv("CHECKLY_API_KEY")
	if key == "" {
		t.Fatal("'CHECKLY_API_KEY' must be set for integration tests")
	}
	return key
}

func TestCreateIntegration(t *testing.T) {
	t.Parallel()
	client := NewClient(getAPIKey(t))
	checkCreate := Check{
		Name:      "integrationTestCreate",
		Type:      TypeBrowser,
		Activated: true,
	}
	ID, err := client.Create(checkCreate)
	defer client.Delete(ID)
	if err != nil {
		t.Fatal(err)
	}
	if !idRE.MatchString(ID) {
		t.Errorf("malformed ID %q (should match %q)", ID, idFormat)
	}
}

func TestGetIntegration(t *testing.T) {
	t.Parallel()
	client := NewClient(getAPIKey(t))
	checkCreate := Check{
		Name:      "integrationTestGet",
		Type:      TypeBrowser,
		Activated: true,
	}
	ID, err := client.Create(checkCreate)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Delete(ID)
	check, err := client.Get(ID)
	if err != nil {
		t.Fatal(err)
	}
	if check.Name != "integrationTestGet" {
		t.Errorf("want 'integrationTestGet', got %q", check.Name)
	}
}

func TestDeleteIntegration(t *testing.T) {
	t.Parallel()
	client := NewClient(getAPIKey(t))
	checkCreate := Check{
		Name:      "integrationTestDelete",
		Type:      TypeBrowser,
		Activated: true,
	}
	ID, err := client.Create(checkCreate)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Delete(ID); err != nil {
		t.Error(err)
	}
}
