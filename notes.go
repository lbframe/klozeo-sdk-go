package klozeo

import "context"

// noteContentBody is the request body for create/update note calls.
type noteContentBody struct {
	Content string `json:"content"`
}

// noteListResponse is the JSON wrapper for the list-notes endpoint.
type noteListResponse struct {
	Notes []*Note `json:"notes"`
}

// CreateNote creates a note on the lead identified by leadID.
func (c *Client) CreateNote(ctx context.Context, leadID, content string) (*Note, error) {
	body := noteContentBody{Content: content}
	var note Note
	if err := c.doJSON(ctx, "POST", "/leads/"+leadID+"/notes", body, &note); err != nil {
		return nil, err
	}
	return &note, nil
}

// ListNotes returns all notes for the lead identified by leadID.
func (c *Client) ListNotes(ctx context.Context, leadID string) ([]*Note, error) {
	var resp noteListResponse
	if err := c.doJSON(ctx, "GET", "/leads/"+leadID+"/notes", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Notes, nil
}

// UpdateNote updates the content of the note identified by noteID.
func (c *Client) UpdateNote(ctx context.Context, noteID, content string) (*Note, error) {
	body := noteContentBody{Content: content}
	var note Note
	if err := c.doJSON(ctx, "PUT", "/notes/"+noteID, body, &note); err != nil {
		return nil, err
	}
	return &note, nil
}

// DeleteNote deletes the note identified by noteID. Returns nil on success.
func (c *Client) DeleteNote(ctx context.Context, noteID string) error {
	return c.doNoContent(ctx, "DELETE", "/notes/"+noteID, nil)
}
