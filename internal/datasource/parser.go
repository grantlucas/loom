package datasource

import (
	"encoding/json"
	"fmt"
)

// ParseIssueList parses the JSON output of bd list --json.
func ParseIssueList(data []byte) ([]Issue, error) {
	var issues []Issue
	if err := json.Unmarshal(data, &issues); err != nil {
		return nil, err
	}
	return issues, nil
}

// ParseIssueDetail parses the JSON output of bd show <id> --json.
// bd show returns a JSON array with a single element.
func ParseIssueDetail(data []byte) (*IssueDetail, error) {
	var details []IssueDetail
	if err := json.Unmarshal(data, &details); err != nil {
		return nil, err
	}
	if len(details) == 0 {
		return nil, fmt.Errorf("empty response from bd show")
	}
	return &details[0], nil
}
