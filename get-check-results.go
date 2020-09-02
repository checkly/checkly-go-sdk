package checkly

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// GetCheckResults gets the results of the given Check
func (c *Client) GetCheckResults(
	checkID string,
	filters *CheckResultsFilter,
) ([]CheckResult, error) {
	uri := fmt.Sprintf("check-results/%s", checkID)
	if filters != nil {
		q := url.Values{}
		q.Add("page", fmt.Sprintf("%d", filters.Page))
		q.Add("limit", fmt.Sprintf("%d", filters.Limit))
		q.Add("to", fmt.Sprintf("%d", filters.To))
		q.Add("from", fmt.Sprintf("%d", filters.From))
		if filters.CheckType == TypeBrowser || filters.CheckType == TypeAPI {
			q.Add("checkType", string(filters.CheckType))
		}
		if filters.HasFailures {
			q.Add("hasFailures", "1")
		}
		uri = uri + "?" + q.Encode()
	}

	status, res, err := c.MakeAPICall(
		http.MethodGet,
		uri,
		nil,
	)

	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status %d: %q", status, res)
	}
	result := []CheckResult{}
	err = json.NewDecoder(strings.NewReader(res)).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("decoding error for data %q: %v", res, err)
	}
	return result, nil
}
