package klozeo

import "context"

// scoringRuleListResponse is the JSON wrapper for the list-scoring-rules endpoint.
type scoringRuleListResponse struct {
	Rules []*ScoringRule `json:"rules"`
}

// ListScoringRules returns all scoring rules for the authenticated account.
func (c *Client) ListScoringRules(ctx context.Context) ([]*ScoringRule, error) {
	var resp scoringRuleListResponse
	if err := c.doJSON(ctx, "GET", "/scoring-rules", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Rules, nil
}

// CreateScoringRule creates a new scoring rule.
func (c *Client) CreateScoringRule(ctx context.Context, rule *ScoringRule) (*ScoringRule, error) {
	var created ScoringRule
	if err := c.doJSON(ctx, "POST", "/scoring-rules", rule, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// GetScoringRule retrieves a single scoring rule by its ID.
func (c *Client) GetScoringRule(ctx context.Context, id string) (*ScoringRule, error) {
	var rule ScoringRule
	if err := c.doJSON(ctx, "GET", "/scoring-rules/"+id, nil, &rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

// UpdateScoringRule applies a partial update to a scoring rule.
// Only non-nil fields in input are sent to the API.
func (c *Client) UpdateScoringRule(ctx context.Context, id string, input *ScoringRuleInput) error {
	return c.doJSON(ctx, "PUT", "/scoring-rules/"+id, input, nil)
}

// DeleteScoringRule deletes a scoring rule by its ID.
func (c *Client) DeleteScoringRule(ctx context.Context, id string) error {
	return c.doJSON(ctx, "DELETE", "/scoring-rules/"+id, nil, nil)
}

// RecalculateScore recalculates and persists the score for the lead identified by leadID.
// It returns the new score value.
func (c *Client) RecalculateScore(ctx context.Context, leadID string) (float64, error) {
	var resp ScoreResponse
	if err := c.doJSON(ctx, "POST", "/leads/"+leadID+"/score", nil, &resp); err != nil {
		return 0, err
	}
	return resp.Score, nil
}
