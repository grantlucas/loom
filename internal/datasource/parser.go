package datasource

import "encoding/json"

// ParseIssueList parses the JSON output of bd list --json.
func ParseIssueList(data []byte) ([]Issue, error) {
	var issues []Issue
	if err := json.Unmarshal(data, &issues); err != nil {
		return nil, err
	}
	return issues, nil
}
