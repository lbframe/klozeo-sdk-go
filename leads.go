package klozeo

import (
	"context"
	"fmt"
	"io"
	"iter"
	"net/url"
)

// Create creates a new lead and returns a CreateResponse.
// If the lead matches an existing one, it is merged (Duplicate=true in the response).
// Use Get to fetch the full lead after creation.
func (c *Client) Create(ctx context.Context, lead *Lead) (*CreateResponse, error) {
	var result CreateResponse
	if err := c.doJSON(ctx, "POST", "/leads", lead, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves a single lead by its ID (cl_<uuid>).
func (c *Client) Get(ctx context.Context, id string) (*Lead, error) {
	var lead Lead
	if err := c.doJSON(ctx, "GET", "/leads/"+id, nil, &lead); err != nil {
		return nil, err
	}
	return &lead, nil
}

// Update applies a partial update to a lead and returns the updated lead.
// Only non-nil fields in input are sent to the API.
func (c *Client) Update(ctx context.Context, id string, input *UpdateLeadInput) (*Lead, error) {
	var lead Lead
	if err := c.doJSON(ctx, "PUT", "/leads/"+id, input, &lead); err != nil {
		return nil, err
	}
	return &lead, nil
}

// Delete deletes a lead by its ID. It returns nil on success (HTTP 204).
func (c *Client) Delete(ctx context.Context, id string) error {
	return c.doNoContent(ctx, "DELETE", "/leads/"+id, nil)
}

// List fetches a single page of leads with the provided options.
// Use Cursor, Sort, Limit, and filter builder functions as options.
//
// Example:
//
//	result, err := client.List(ctx,
//	    klozeo.City().Eq("Berlin"),
//	    klozeo.Sort(klozeo.FieldRating, klozeo.Desc),
//	    klozeo.Limit(20),
//	)
func (c *Client) List(ctx context.Context, opts ...ListOption) (*ListResult, error) {
	lo := buildListOptions(opts)
	q := url.Values{}
	lo.applyToURL(q)
	path := "/leads?" + q.Encode()

	var result ListResult
	if err := c.doJSON(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Iterator returns a Go 1.23 range-over-func iterator that pages through all
// matching leads transparently. Each iteration yields a *Lead and an error.
// On error, iteration stops and the error is yielded.
//
// Example:
//
//	for lead, err := range client.Iterator(ctx, klozeo.City().Eq("Berlin")) {
//	    if err != nil { panic(err) }
//	    fmt.Println(lead.Name)
//	}
func (c *Client) Iterator(ctx context.Context, opts ...ListOption) iter.Seq2[*Lead, error] {
	return func(yield func(*Lead, error) bool) {
		// Copy opts so we can append a cursor without mutating the caller's slice.
		current := make([]ListOption, len(opts))
		copy(current, opts)

		for {
			result, err := c.List(ctx, current...)
			if err != nil {
				yield(nil, err)
				return
			}
			for _, lead := range result.Leads {
				if !yield(lead, nil) {
					return
				}
			}
			if !result.HasMore || result.NextCursor == "" {
				return
			}
			// Replace any existing cursor option or append a new one.
			current = replaceOrAppendCursor(current, result.NextCursor)
		}
	}
}

// IteratorChan returns a channel of leads and a channel of errors.
// Both channels are closed when iteration completes or a fatal error occurs.
// The caller should consume both channels concurrently.
//
// Example:
//
//	leads, errs := client.IteratorChan(ctx, klozeo.City().Eq("Berlin"))
//	go func() {
//	    for err := range errs { log.Println(err) }
//	}()
//	for lead := range leads {
//	    process(lead)
//	}
func (c *Client) IteratorChan(ctx context.Context, opts ...ListOption) (<-chan *Lead, <-chan error) {
	leads := make(chan *Lead)
	errs := make(chan error, 1)

	go func() {
		defer close(leads)
		defer close(errs)

		for lead, err := range c.Iterator(ctx, opts...) {
			if err != nil {
				errs <- err
				return
			}
			select {
			case leads <- lead:
			case <-ctx.Done():
				errs <- ctx.Err()
				return
			}
		}
	}()

	return leads, errs
}

// BatchCreate creates up to 100 leads in a single request.
// Returns a BatchResult with per-item outcomes.
func (c *Client) BatchCreate(ctx context.Context, leads []*Lead) (*BatchResult, error) {
	body := map[string]any{"leads": leads}
	var result BatchResult
	if err := c.doJSON(ctx, "POST", "/leads/batch", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BatchUpdate applies the same partial update to multiple leads.
// ids must not exceed 100 entries.
func (c *Client) BatchUpdate(ctx context.Context, ids []string, input *UpdateLeadInput) (*BatchResult, error) {
	body := map[string]any{"ids": ids, "data": input}
	var result BatchResult
	if err := c.doJSON(ctx, "PUT", "/leads/batch", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BatchDelete deletes multiple leads by their IDs.
// ids must not exceed 100 entries.
func (c *Client) BatchDelete(ctx context.Context, ids []string) (*BatchResult, error) {
	body := map[string]any{"ids": ids}
	var result BatchResult
	if err := c.doJSON(ctx, "DELETE", "/leads/batch", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BatchCreateFromChan reads leads from ch, groups them into batches of 100,
// and creates each batch via BatchCreate. It returns a channel of *CreateResponse
// (one entry per successfully created lead) and an error channel.
// Both channels are closed when the input channel is drained or a fatal error occurs.
//
// Example:
//
//	leads := make(chan *klozeo.Lead)
//	go func() { defer close(leads); /* send leads */ }()
//	results, errs := client.BatchCreateFromChan(ctx, leads)
//	go func() { for err := range errs { log.Println(err) } }()
//	for r := range results { fmt.Println(r.ID) }
func (c *Client) BatchCreateFromChan(ctx context.Context, ch <-chan *Lead) (<-chan *CreateResponse, <-chan error) {
	out := make(chan *CreateResponse)
	errs := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errs)

		const batchSize = 100
		batch := make([]*Lead, 0, batchSize)

		flush := func() bool {
			if len(batch) == 0 {
				return true
			}
			result, err := c.BatchCreate(ctx, batch)
			batch = batch[:0]
			if err != nil {
				errs <- fmt.Errorf("klozeo: batch create: %w", err)
				return false
			}
			for _, item := range result.Created {
				resp := &CreateResponse{
					ID:        item.ID,
					CreatedAt: item.CreatedAt,
				}
				select {
				case out <- resp:
				case <-ctx.Done():
					errs <- ctx.Err()
					return false
				}
			}
			return true
		}

		for {
			select {
			case lead, ok := <-ch:
				if !ok {
					flush()
					return
				}
				batch = append(batch, lead)
				if len(batch) >= batchSize {
					if !flush() {
						return
					}
				}
			case <-ctx.Done():
				errs <- ctx.Err()
				return
			}
		}
	}()

	return out, errs
}

// Export streams all matching leads in the specified format.
// The returned io.ReadCloser must be closed by the caller.
//
// Example:
//
//	reader, err := client.Export(ctx, klozeo.ExportCSV, klozeo.City().Eq("Paris"))
//	defer reader.Close()
//	io.Copy(file, reader)
func (c *Client) Export(ctx context.Context, format ExportFormat, opts ...ListOption) (io.ReadCloser, error) {
	lo := buildListOptions(opts)
	q := url.Values{}
	lo.applyToURL(q)
	q.Set("format", string(format))
	path := "/leads/export?" + q.Encode()

	resp, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// replaceOrAppendCursor updates or appends a Cursor option in a slice of ListOptions.
func replaceOrAppendCursor(opts []ListOption, cursor string) []ListOption {
	result := make([]ListOption, 0, len(opts)+1)
	replaced := false
	for _, o := range opts {
		if _, ok := o.(cursorOption); ok {
			result = append(result, Cursor(cursor))
			replaced = true
		} else {
			result = append(result, o)
		}
	}
	if !replaced {
		result = append(result, Cursor(cursor))
	}
	return result
}
