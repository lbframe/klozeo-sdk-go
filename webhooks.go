package klozeo

import "context"

// webhookListResponse is the JSON wrapper for the list-webhooks endpoint.
type webhookListResponse struct {
	Webhooks []*Webhook `json:"webhooks"`
}

// ListWebhooks returns all webhook subscriptions for the authenticated account.
func (c *Client) ListWebhooks(ctx context.Context) ([]*Webhook, error) {
	var resp webhookListResponse
	if err := c.doJSON(ctx, "GET", "/webhooks", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Webhooks, nil
}

// CreateWebhook creates a new webhook subscription.
// input.URL is required; Events and Secret are optional.
func (c *Client) CreateWebhook(ctx context.Context, input *WebhookInput) (*Webhook, error) {
	var webhook Webhook
	if err := c.doJSON(ctx, "POST", "/webhooks", input, &webhook); err != nil {
		return nil, err
	}
	return &webhook, nil
}

// DeleteWebhook deletes the webhook identified by id.
func (c *Client) DeleteWebhook(ctx context.Context, id string) error {
	return c.doJSON(ctx, "DELETE", "/webhooks/"+id, nil, nil)
}
