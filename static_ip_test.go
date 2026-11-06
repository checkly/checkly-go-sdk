package checkly_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	checkly "github.com/checkly/checkly-go-sdk"
)

func TestGetStaticIPs(t *testing.T) {
	t.Parallel()

	fixtureMap := map[string]string{
		"/v1/static-ips-by-region":   "StaticIPs.json",
		"/v1/static-ipv6s-by-region": "StaticIPv6s.json",
	}

	// we can't use cannedResponseServer here since we need a response on more than one URL
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fileName, ok := fixtureMap[r.URL.Path]
		if !ok {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}

		filePath := filepath.Join("fixtures", fileName)
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		_, _ = w.Write(fileData)
	}))
	defer ts.Close()

	client := newTestClient(t, checkly.WithBaseURL(ts.URL), checkly.WithAuth("dummy-key", testAccountID), checkly.WithHTTPClient(ts.Client()))
	gotStaticIPs, err := client.GetStaticIPs(context.Background())
	if err != nil {
		t.Error(err)
	}

	exampleIPv4 := netip.MustParsePrefix("54.151.146.209/32")
	exampleIPv6 := netip.MustParsePrefix("2600:1f18:12ca:3000::/56")

	expected := []checkly.StaticIP{
		{Region: "ap-southeast-1", Address: exampleIPv4},
		{Region: "us-east-1", Address: exampleIPv6},
	}

	for _, exp := range expected {
		found := false
		for _, ip := range gotStaticIPs {
			if reflect.DeepEqual(ip, exp) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %+v to be included in %+v, but it was not found", exp, gotStaticIPs)
		}
	}
}
