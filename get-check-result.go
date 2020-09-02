package checkly

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// GetCheckResult gets a specific Check result
func (c *Client) GetCheckResult(checkID, checkResultID string) (CheckResult, error) {
	status, res, err := c.MakeAPICall(
		http.MethodGet,
		fmt.Sprintf("check-results/%s/%s", checkID, checkResultID),
		nil,
	)
	result := CheckResult{}
	if err != nil {
		return result, err
	}
	if status != http.StatusOK {
		return result, fmt.Errorf("unexpected response status %d: %q", status, res)
	}

	err = json.NewDecoder(strings.NewReader(res)).Decode(&result)
	if err != nil {
		return result, fmt.Errorf("decoding error for data %q: %v", res, err)
	}
	return result, nil
}
