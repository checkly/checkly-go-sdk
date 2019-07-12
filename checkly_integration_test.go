// +build integration

package checkly

import (
	"net/http"
	"os"
	"reflect"
	"testing"
)

func getAPIKey(t *testing.T) string {
	key := os.Getenv("CHECKLY_API_KEY")
	if key == "" {
		t.Fatal("'CHECKLY_API_KEY' must be set for integration tests")
	}
	return key
}

func testCheck(name string) Check {
	return Check{
		Name:      name,
		Type:      TypeAPI,
		Activated: true,
		Request: Request{
			Method: http.MethodGet,
			URL:    "http://example.com",
		},
	}
}
func TestCreateGetIntegration(t *testing.T) {
	t.Parallel()
	client := NewClient(getAPIKey(t))
	checkCreate := testCheck("integrationTestCreate")
	ID, err := client.Create(checkCreate)
	// defer client.Delete(ID)
	if err != nil {
		t.Fatal(err)
	}
	check, err := client.Get(ID)
	checkCreate.ID = ID
	if !reflect.DeepEqual(checkCreate, check) {
		t.Errorf("mismatch: want %+v, got %+v", checkCreate, check)
	}
}

func TestDeleteIntegration(t *testing.T) {
	t.Parallel()
	client := NewClient(getAPIKey(t))
	checkDelete := testCheck("integrationTestDelete")
	ID, err := client.Create(checkDelete)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Delete(ID); err != nil {
		t.Error(err)
	}
	_, err = client.Get(ID)
	if err == nil {
		t.Error("want error getting deleted check, but got nil")
	}
}
