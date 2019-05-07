package checkly

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCreateCheck(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := os.Open("testdata/Check.json")
		if err != nil {
			t.Fatal(err)
		}
		io.Copy(w, data)
	}))
	defer ts.Close()
	client := NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	// client.Debug = os.Stdout
	check, err := client.CreateCheck(Params{"name": "testCheck"})
	if err != nil {
		t.Fatal(err)
	}
	wantName := "testCheck"
	if check.Name != wantName {
		t.Fatalf("want %q, got %q", wantName, check.Name)
	}
}
