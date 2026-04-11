// Package klozeo provides a Go client for the Klozeo API.
// It supports creating, reading, updating, and deleting leads,
// notes, scoring rules, and webhooks, with cursor-based pagination,
// filter builders, and automatic retry with exponential backoff.
package klozeo

// Lead represents a lead record in Klozeo.
type Lead struct {
	// ID is the unique identifier of the lead (cl_<uuid>). Read-only on create.
	ID string `json:"id,omitempty"`
	// Score is the computed lead score (0–100). Read-only.
	Score float64 `json:"score,omitempty"`
	// CreatedAt is the Unix timestamp (seconds) when the lead was created.
	CreatedAt int64 `json:"created_at,omitempty"`
	// UpdatedAt is the Unix timestamp (seconds) when the lead was last structurally updated.
	UpdatedAt int64 `json:"updated_at,omitempty"`
	// LastInteractionAt is the Unix timestamp (seconds) of the last inbound push or merge.
	LastInteractionAt int64 `json:"last_interaction_at,omitempty"`

	// Name is the lead name. Required.
	Name string `json:"name"`
	// Source is the lead source. Required.
	Source string `json:"source"`

	// Description is an optional description.
	Description string `json:"description,omitempty"`
	// Address is the street address.
	Address string `json:"address,omitempty"`
	// City is the city.
	City string `json:"city,omitempty"`
	// State is the state or region.
	State string `json:"state,omitempty"`
	// Country is the country.
	Country string `json:"country,omitempty"`
	// PostalCode is the postal or ZIP code.
	PostalCode string `json:"postal_code,omitempty"`
	// Latitude is the geographic latitude.
	Latitude *float64 `json:"latitude,omitempty"`
	// Longitude is the geographic longitude.
	Longitude *float64 `json:"longitude,omitempty"`
	// Phone is the phone number.
	Phone string `json:"phone,omitempty"`
	// Email is the email address.
	Email string `json:"email,omitempty"`
	// Website is the website URL.
	Website string `json:"website,omitempty"`
	// Rating is the numeric rating (0–5).
	Rating *float64 `json:"rating,omitempty"`
	// ReviewCount is the number of reviews.
	ReviewCount *int `json:"review_count,omitempty"`
	// Category is the business category.
	Category string `json:"category,omitempty"`
	// Tags is a list of string tags.
	Tags []string `json:"tags,omitempty"`
	// SourceID is an external identifier from the originating source.
	SourceID string `json:"source_id,omitempty"`
	// LogoURL is the URL of the lead's logo image.
	LogoURL string `json:"logo_url,omitempty"`

	// Attributes is the list of dynamic custom attributes.
	Attributes []Attribute `json:"attributes,omitempty"`
}

// UpdateLeadInput holds the fields for a partial lead update.
// Only non-nil pointer fields are included in the request body.
type UpdateLeadInput struct {
	// Name updates the lead name.
	Name *string `json:"name,omitempty"`
	// Source updates the lead source.
	Source *string `json:"source,omitempty"`
	// Description updates the description.
	Description *string `json:"description,omitempty"`
	// Address updates the street address.
	Address *string `json:"address,omitempty"`
	// City updates the city.
	City *string `json:"city,omitempty"`
	// State updates the state or region.
	State *string `json:"state,omitempty"`
	// Country updates the country.
	Country *string `json:"country,omitempty"`
	// PostalCode updates the postal code.
	PostalCode *string `json:"postal_code,omitempty"`
	// Latitude updates the geographic latitude.
	Latitude *float64 `json:"latitude,omitempty"`
	// Longitude updates the geographic longitude.
	Longitude *float64 `json:"longitude,omitempty"`
	// Phone updates the phone number.
	Phone *string `json:"phone,omitempty"`
	// Email updates the email address.
	Email *string `json:"email,omitempty"`
	// Website updates the website URL.
	Website *string `json:"website,omitempty"`
	// Rating updates the numeric rating.
	Rating *float64 `json:"rating,omitempty"`
	// ReviewCount updates the review count.
	ReviewCount *int `json:"review_count,omitempty"`
	// Category updates the business category.
	Category *string `json:"category,omitempty"`
	// Tags replaces the tags list.
	Tags []string `json:"tags,omitempty"`
	// SourceID updates the external source identifier.
	SourceID *string `json:"source_id,omitempty"`
	// LogoURL updates the logo image URL.
	LogoURL *string `json:"logo_url,omitempty"`
}

// Attribute represents a dynamic custom attribute on a lead.
type Attribute struct {
	// ID is the unique attribute identifier. Populated in API responses.
	ID string `json:"id,omitempty"`
	// Name is the attribute name.
	Name string `json:"name"`
	// Type is the attribute type: "text", "number", "bool", "list", or "object".
	Type string `json:"type"`
	// Value is the attribute value. Its underlying type depends on Type.
	Value any `json:"value"`
}

// TextAttr creates a text attribute.
func TextAttr(name, value string) Attribute {
	return Attribute{Name: name, Type: "text", Value: value}
}

// NumberAttr creates a number attribute.
func NumberAttr(name string, value float64) Attribute {
	return Attribute{Name: name, Type: "number", Value: value}
}

// BoolAttr creates a boolean attribute.
func BoolAttr(name string, value bool) Attribute {
	return Attribute{Name: name, Type: "bool", Value: value}
}

// ListAttr creates a list attribute.
func ListAttr(name string, value []string) Attribute {
	return Attribute{Name: name, Type: "list", Value: value}
}

// ObjectAttr creates an object attribute.
func ObjectAttr(name string, value map[string]any) Attribute {
	return Attribute{Name: name, Type: "object", Value: value}
}

// Note represents a note attached to a lead.
type Note struct {
	// ID is the note identifier (note_<uuid>).
	ID string `json:"id"`
	// LeadID is the parent lead identifier (cl_<uuid>).
	LeadID string `json:"lead_id"`
	// Content is the note text content.
	Content string `json:"content"`
	// CreatedAt is the Unix timestamp (seconds) when the note was created.
	CreatedAt int64 `json:"created_at"`
	// UpdatedAt is the Unix timestamp (seconds) when the note was last updated.
	UpdatedAt int64 `json:"updated_at"`
}

// CreateResponse is returned by the Create endpoint.
// Use Get to retrieve the full lead after creation.
type CreateResponse struct {
	// ID is the lead identifier (cl_<uuid>).
	ID string `json:"id"`
	// Message is a human-readable status message.
	Message string `json:"message"`
	// CreatedAt is the Unix timestamp (seconds) of creation or merge.
	CreatedAt int64 `json:"created_at"`
	// Duplicate is true when the incoming lead was merged into an existing one.
	Duplicate bool `json:"duplicate,omitempty"`
	// PotentialDuplicateID is set when a low-confidence duplicate was detected.
	PotentialDuplicateID string `json:"potential_duplicate_id,omitempty"`
}

// ListResult is the response from the List endpoint.
type ListResult struct {
	// Leads is the page of leads.
	Leads []*Lead `json:"leads"`
	// NextCursor is an opaque token for fetching the next page.
	NextCursor string `json:"next_cursor"`
	// HasMore indicates whether more results exist beyond this page.
	HasMore bool `json:"has_more"`
	// Count is the number of leads in this page.
	Count int `json:"count"`
}

// BatchCreatedItem represents a successfully created lead in a batch response.
type BatchCreatedItem struct {
	// Index is the zero-based position in the input slice.
	Index int `json:"index"`
	// ID is the created lead identifier.
	ID string `json:"id"`
	// CreatedAt is the Unix timestamp of creation.
	CreatedAt int64 `json:"created_at"`
}

// BatchError represents a failure for a single item in a batch operation.
type BatchError struct {
	// Index is the zero-based position in the input slice.
	Index int `json:"index"`
	// Message describes why the item failed.
	Message string `json:"message"`
}

// BatchResultItem represents a single result in a batch update or delete response.
type BatchResultItem struct {
	// Index is the zero-based position in the input IDs slice.
	Index int `json:"index"`
	// ID is the lead identifier.
	ID string `json:"id"`
	// Success indicates whether the operation succeeded for this item.
	Success bool `json:"success"`
	// Message is set when Success is false.
	Message string `json:"message,omitempty"`
}

// BatchResult is the response from batch create, update, or delete operations.
type BatchResult struct {
	// Created contains successfully created items (batch create only).
	Created []*BatchCreatedItem `json:"created,omitempty"`
	// Results contains per-item outcomes (batch update/delete only).
	Results []*BatchResultItem `json:"results,omitempty"`
	// Errors contains items that failed during batch create.
	Errors []*BatchError `json:"errors,omitempty"`
	// Total is the total number of items in the request.
	Total int `json:"total"`
	// Success is the number of items that succeeded.
	Success int `json:"success"`
	// Failed is the number of items that failed.
	Failed int `json:"failed"`
}

// ScoringRule represents a named scoring rule with an expression.
type ScoringRule struct {
	// ID is the unique identifier of the rule (raw UUID).
	ID string `json:"id,omitempty"`
	// Name is the human-readable rule name.
	Name string `json:"name"`
	// Expression is the scoring expression string.
	Expression string `json:"expression"`
	// Priority controls evaluation order; lower value = higher priority.
	Priority int `json:"priority,omitempty"`
	// CreatedAt is the Unix timestamp of creation.
	CreatedAt int64 `json:"created_at,omitempty"`
	// UpdatedAt is the Unix timestamp of the last update.
	UpdatedAt int64 `json:"updated_at,omitempty"`
}

// ScoringRuleInput holds fields for a partial scoring rule update.
type ScoringRuleInput struct {
	// Name updates the rule name.
	Name *string `json:"name,omitempty"`
	// Expression updates the scoring expression.
	Expression *string `json:"expression,omitempty"`
	// Priority updates the evaluation priority.
	Priority *int `json:"priority,omitempty"`
}

// Webhook represents an outbound webhook subscription.
type Webhook struct {
	// ID is the unique identifier of the webhook (raw UUID).
	ID string `json:"id"`
	// URL is the endpoint that receives event payloads.
	URL string `json:"url"`
	// Events is the list of event names to subscribe to.
	Events []string `json:"events"`
	// Active indicates whether the webhook is enabled.
	Active bool `json:"active"`
	// CreatedAt is the ISO 8601 timestamp of creation.
	CreatedAt string `json:"created_at"`
}

// WebhookInput holds the fields for creating a webhook.
type WebhookInput struct {
	// URL is the endpoint that receives event payloads. Required.
	URL string `json:"url"`
	// Events is the list of event names. Optional.
	Events []string `json:"events,omitempty"`
	// Secret is an optional signing secret for payload verification.
	// It is never returned by the API after creation.
	Secret string `json:"secret,omitempty"`
}

// ScoreResponse is returned by the RecalculateScore endpoint.
type ScoreResponse struct {
	// ID is the lead identifier.
	ID string `json:"id"`
	// Score is the newly computed score.
	Score float64 `json:"score"`
}

// RateLimitState holds the last observed rate limit headers from the API.
type RateLimitState struct {
	// Limit is the maximum number of requests allowed per window.
	Limit int
	// Remaining is the number of requests remaining in the current window.
	Remaining int
}

// ExportFormat specifies the file format for lead exports.
type ExportFormat string

const (
	// ExportCSV exports leads as a CSV file.
	ExportCSV ExportFormat = "csv"
	// ExportJSON exports leads as a JSON file.
	ExportJSON ExportFormat = "json"
	// ExportXLSX exports leads as an Excel file.
	ExportXLSX ExportFormat = "xlsx"
)

// Sort order constants for use with the Sort list option.
const (
	// Asc sorts in ascending order.
	Asc = "ASC"
	// Desc sorts in descending order.
	Desc = "DESC"
)

// Sortable field name constants.
const (
	FieldName              = "name"
	FieldCity              = "city"
	FieldCountry           = "country"
	FieldState             = "state"
	FieldCategory          = "category"
	FieldSource            = "source"
	FieldEmail             = "email"
	FieldPhone             = "phone"
	FieldWebsite           = "website"
	FieldRating            = "rating"
	FieldReviewCount       = "review_count"
	FieldCreatedAt         = "created_at"
	FieldUpdatedAt         = "updated_at"
	FieldLastInteractionAt = "last_interaction_at"
)

// AttrSortField returns the sort field name for a custom attribute.
func AttrSortField(name string) string {
	return "attr:" + name
}

// Ptr returns a pointer to v. It is a convenience helper for optional fields.
func Ptr[T any](v T) *T {
	return &v
}
