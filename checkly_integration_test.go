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
	params := Params{
		"name":      "integrationTestCreate",
		"checkType": "BROWSER",
		"activated": "true",
	}
	ID, err := client.Create(params)
	defer client.Delete(ID)
	if err != nil {
		t.Fatal(err)
	}
	if !idRE.MatchString(ID) {
		t.Errorf("malformed ID %q (should match %q)", ID, idFormat)
	}
}

func TestDeleteIntegration(t *testing.T) {
	t.Parallel()
	client := NewClient(getAPIKey(t))
	params := Params{
		"name":      "integrationTestDelete",
		"checkType": "BROWSER",
		"activated": "true",
	}
	ID, err := client.Create(params)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Delete(ID); err != nil {
		t.Error(err)
	}
}