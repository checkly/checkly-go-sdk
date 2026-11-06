package checkly_test

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"testing"

	checkly "github.com/checkly/checkly-go-sdk"
)

// The assertion lists are replaced on every update, so a regex must be sent
// with its assertion to survive the update.
func TestAssertionRegexPayload(t *testing.T) {
	t.Parallel()
	c, lastBody := newRecordingServer(t, http.StatusOK, `{"checkType":"API"}`)
	_, err := c.UpdateAPICheck(context.Background(), "id", checkly.APICheck{
		Name: "n",
		Request: checkly.Request{Assertions: []checkly.Assertion{
			{Source: checkly.AssertionSourceHeaders, Property: "Cache-Control", Comparison: checkly.AssertionComparisonEquals, Target: "60", Regex: `max-age=(\d+)`},
			{Source: checkly.AssertionSourceStatusCode, Comparison: checkly.AssertionComparisonEquals, Target: "200"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var body struct {
		Request struct {
			Assertions []map[string]json.RawMessage `json:"assertions"`
		} `json:"request"`
	}
	if err := json.Unmarshal(lastBody(), &body); err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, a := range body.Request.Assertions {
		got = append(got, string(a["regex"]))
	}
	if want := []string{`"max-age=(\\d+)"`, `""`}; !slices.Equal(want, got) {
		t.Errorf("want regexes %v, got %v", want, got)
	}
}

func TestAssertionRegexDecoding(t *testing.T) {
	t.Parallel()
	c, _ := newRecordingServer(t, http.StatusOK, `{"checkType":"API","request":{"assertions":[{"source":"HEADERS","regex":"a(b)"},{"source":"STATUS_CODE","regex":null}]}}`)
	check, err := c.GetAPICheck(context.Background(), "id")
	if err != nil {
		t.Fatal(err)
	}
	if got := check.Request.Assertions; len(got) != 2 || got[0].Regex != "a(b)" || got[1].Regex != "" {
		t.Errorf(`want regexes "a(b)" and "", got %+v`, got)
	}
}
