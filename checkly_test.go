package checkly

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func assertFormParamPresent(t *testing.T, form url.Values, param string) {
	value := form.Get(param)
	if value == "" {
		t.Errorf("want %q parameter, got none", param)
	}
}
func TestCreateCheck(t *testing.T) {
	t.Parallel()
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("want POST request, got %q", r.Method)
		}
		wantURL := "/v1/checks"
		if r.URL.EscapedPath() != wantURL {
			t.Errorf("want %q, got %q", wantURL, r.URL.EscapedPath())
		}
		r.ParseForm()
		wantParams := []string{"name", "checkType", "activated"}
		for _, p := range wantParams {
			assertFormParamPresent(t, r.Form, p)
		}
		data, err := os.Open("testdata/Check.json")
		if err != nil {
			t.Fatal(err)
		}
		defer data.Close()
		io.Copy(w, data)
	}))
	defer ts.Close()
	client := NewClient("dummy")
	client.HTTPClient = ts.Client()
	client.URL = ts.URL
	// client.Debug = os.Stdout
	wantName := "test"
	params := Params{
		"name":      wantName,
		"checkType": "BROWSER",
		"activated": "true",
	}
	check, err := client.CreateCheck(params)
	if err != nil {
		t.Fatal(err)
	}
	if check.Name != wantName {
		t.Fatalf("want %q, got %q", wantName, check.Name)
	}
}

func TestDeleteCheck(t *testing.T) {
	t.Parallel()
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
	wantName := "test"
	check, err := client.CreateCheck(Params{"name": wantName})
	if err != nil {
		t.Fatal(err)
	}
	if check.Name != wantName {
		t.Fatalf("want %q, got %q", wantName, check.Name)
	}
}
