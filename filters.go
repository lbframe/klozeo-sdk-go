package klozeo

import (
	"fmt"
	"net/url"
	"strconv"
)

// listOptions collects all options passed to List, Iterator, IteratorChan, and Export.
type listOptions struct {
	filters []filterParam
	sortBy  string
	sortOrd string
	limit   int
	cursor  string
}

// filterParam represents a single serialized filter query parameter value.
type filterParam struct {
	value string // e.g. "and.eq.city.Berlin"
}

// ListOption is a functional option for list-style API calls.
// Filter builders, Sort, Limit, and Cursor all implement this interface.
type ListOption interface {
	apply(*listOptions)
}

// applyToURL serializes listOptions onto the provided url.Values.
func (o *listOptions) applyToURL(q url.Values) {
	for _, f := range o.filters {
		q.Add("filter", f.value)
	}
	if o.sortBy != "" {
		q.Set("sort_by", o.sortBy)
	}
	if o.sortOrd != "" {
		q.Set("sort_order", o.sortOrd)
	}
	if o.limit > 0 {
		q.Set("limit", strconv.Itoa(o.limit))
	}
	if o.cursor != "" {
		q.Set("cursor", o.cursor)
	}
}

// buildListOptions merges all provided ListOptions into a single listOptions struct.
func buildListOptions(opts []ListOption) listOptions {
	var lo listOptions
	for _, o := range opts {
		o.apply(&lo)
	}
	return lo
}

// --- Sort, Limit, Cursor ---

type sortOption struct {
	field string
	order string
}

func (s sortOption) apply(o *listOptions) {
	o.sortBy = s.field
	o.sortOrd = s.order
}

// Sort returns a ListOption that sets the sort field and order (Asc or Desc).
func Sort(field, order string) ListOption {
	return sortOption{field: field, order: order}
}

type limitOption int

func (l limitOption) apply(o *listOptions) {
	o.limit = int(l)
}

// Limit returns a ListOption that sets the maximum number of results per page.
func Limit(n int) ListOption {
	return limitOption(n)
}

type cursorOption string

func (c cursorOption) apply(o *listOptions) {
	o.cursor = string(c)
}

// Cursor returns a ListOption that sets the pagination cursor from a previous response.
func Cursor(cursor string) ListOption {
	return cursorOption(cursor)
}

// --- Filter builders ---

// logic values for filter expressions.
const (
	logicAnd = "and"
	logicOr  = "or"
)

// filterBase is the common builder for all filter types.
// It holds the logic prefix and the field name.
type filterBase struct {
	logic string
	field string
}

func (b filterBase) makeParam(operator, value string) filterParam {
	return filterParam{value: fmt.Sprintf("%s.%s.%s.%s", b.logic, operator, b.field, value)}
}

func (b filterBase) makeParamNoValue(operator string) filterParam {
	return filterParam{value: fmt.Sprintf("%s.%s.%s", b.logic, operator, b.field)}
}

// --- Text filter ---

// TextFilter is the filter builder for text-type fields.
type TextFilter struct {
	filterBase
}

func (f TextFilter) apply(o *listOptions) { /* not a ListOption directly */ }

// Eq adds an equality filter (case insensitive).
func (f TextFilter) Eq(value string) ListOption {
	return f.makeParam("eq", value)
}

// Neq adds a not-equals filter.
func (f TextFilter) Neq(value string) ListOption {
	return f.makeParam("neq", value)
}

// Contains adds a substring contains filter.
func (f TextFilter) Contains(value string) ListOption {
	return f.makeParam("contains", value)
}

// NotContains adds a not-contains filter.
func (f TextFilter) NotContains(value string) ListOption {
	return f.makeParam("not_contains", value)
}

// IsEmpty adds an is-empty filter (null or empty string).
func (f TextFilter) IsEmpty() ListOption {
	return f.makeParamNoValue("is_empty")
}

// IsNotEmpty adds an is-not-empty filter.
func (f TextFilter) IsNotEmpty() ListOption {
	return f.makeParamNoValue("is_not_empty")
}

// --- Number filter ---

// NumberFilter is the filter builder for numeric fields.
type NumberFilter struct {
	filterBase
}

// Eq adds an equality filter for a number field.
func (f NumberFilter) Eq(value float64) ListOption {
	return f.makeParam("eq", strconv.FormatFloat(value, 'f', -1, 64))
}

// Neq adds a not-equals filter for a number field.
func (f NumberFilter) Neq(value float64) ListOption {
	return f.makeParam("neq", strconv.FormatFloat(value, 'f', -1, 64))
}

// Gt adds a greater-than filter.
func (f NumberFilter) Gt(value float64) ListOption {
	return f.makeParam("gt", strconv.FormatFloat(value, 'f', -1, 64))
}

// Gte adds a greater-than-or-equal filter.
func (f NumberFilter) Gte(value float64) ListOption {
	return f.makeParam("gte", strconv.FormatFloat(value, 'f', -1, 64))
}

// Lt adds a less-than filter.
func (f NumberFilter) Lt(value float64) ListOption {
	return f.makeParam("lt", strconv.FormatFloat(value, 'f', -1, 64))
}

// Lte adds a less-than-or-equal filter.
func (f NumberFilter) Lte(value float64) ListOption {
	return f.makeParam("lte", strconv.FormatFloat(value, 'f', -1, 64))
}

// --- Array (tags) filter ---

// ArrayFilter is the filter builder for array-type fields (e.g. tags).
type ArrayFilter struct {
	filterBase
}

// Contains adds an array-contains filter.
func (f ArrayFilter) Contains(value string) ListOption {
	return f.makeParam("array_contains", value)
}

// NotContains adds an array-not-contains filter.
func (f ArrayFilter) NotContains(value string) ListOption {
	return f.makeParam("array_not_contains", value)
}

// IsEmpty adds an array-is-empty filter.
func (f ArrayFilter) IsEmpty() ListOption {
	return f.makeParamNoValue("array_empty")
}

// IsNotEmpty adds an array-is-not-empty filter.
func (f ArrayFilter) IsNotEmpty() ListOption {
	return f.makeParamNoValue("array_not_empty")
}

// --- Location filter ---

// LocationFilter is the filter builder for the location (lat/lng) field.
type LocationFilter struct {
	filterBase
}

// WithinRadius adds a within-radius filter. lat and lng are in decimal degrees; km is the radius.
func (f LocationFilter) WithinRadius(lat, lng, km float64) ListOption {
	v := fmt.Sprintf("%s,%s,%s",
		strconv.FormatFloat(lat, 'f', -1, 64),
		strconv.FormatFloat(lng, 'f', -1, 64),
		strconv.FormatFloat(km, 'f', -1, 64),
	)
	return f.makeParam("within_radius", v)
}

// IsSet adds a filter that checks coordinates are present.
func (f LocationFilter) IsSet() ListOption {
	return f.makeParamNoValue("is_set")
}

// IsNotSet adds a filter that checks coordinates are absent.
func (f LocationFilter) IsNotSet() ListOption {
	return f.makeParamNoValue("is_not_set")
}

// --- Attribute filter ---

// AttrFilter is the filter builder for custom (dynamic) attributes.
type AttrFilter struct {
	filterBase
}

// Eq adds a text equality filter for a custom attribute.
func (f AttrFilter) Eq(value string) ListOption {
	return f.makeParam("eq", value)
}

// Neq adds a text not-equals filter for a custom attribute.
func (f AttrFilter) Neq(value string) ListOption {
	return f.makeParam("neq", value)
}

// Contains adds a text contains filter for a custom attribute.
func (f AttrFilter) Contains(value string) ListOption {
	return f.makeParam("contains", value)
}

// EqNumber adds a numeric equality filter for a custom attribute.
func (f AttrFilter) EqNumber(value float64) ListOption {
	return f.makeParam("eq", strconv.FormatFloat(value, 'f', -1, 64))
}

// Gt adds a numeric greater-than filter for a custom attribute.
func (f AttrFilter) Gt(value float64) ListOption {
	return f.makeParam("gt", strconv.FormatFloat(value, 'f', -1, 64))
}

// Gte adds a numeric greater-than-or-equal filter for a custom attribute.
func (f AttrFilter) Gte(value float64) ListOption {
	return f.makeParam("gte", strconv.FormatFloat(value, 'f', -1, 64))
}

// Lt adds a numeric less-than filter for a custom attribute.
func (f AttrFilter) Lt(value float64) ListOption {
	return f.makeParam("lt", strconv.FormatFloat(value, 'f', -1, 64))
}

// Lte adds a numeric less-than-or-equal filter for a custom attribute.
func (f AttrFilter) Lte(value float64) ListOption {
	return f.makeParam("lte", strconv.FormatFloat(value, 'f', -1, 64))
}

// --- filterParam as ListOption ---

func (p filterParam) apply(o *listOptions) {
	o.filters = append(o.filters, p)
}

// --- Or modifier ---

// orModifier wraps a logic prefix of "or" and exposes the same filter-builder methods
// so callers can write Or().City().Eq("Paris").
type orModifier struct{}

// Or returns an or-logic modifier. Chain field selectors after it:
//
//	klozeo.Or().City().Eq("Paris")
func Or() *orModifier {
	return &orModifier{}
}

func (m *orModifier) base(field string) filterBase {
	return filterBase{logic: logicOr, field: field}
}

// City returns an OR text filter for the city field.
func (m *orModifier) City() TextFilter { return TextFilter{m.base("city")} }

// Name returns an OR text filter for the name field.
func (m *orModifier) Name() TextFilter { return TextFilter{m.base("name")} }

// Country returns an OR text filter for the country field.
func (m *orModifier) Country() TextFilter { return TextFilter{m.base("country")} }

// State returns an OR text filter for the state field.
func (m *orModifier) State() TextFilter { return TextFilter{m.base("state")} }

// Category returns an OR text filter for the category field.
func (m *orModifier) Category() TextFilter { return TextFilter{m.base("category")} }

// Source returns an OR text filter for the source field.
func (m *orModifier) Source() TextFilter { return TextFilter{m.base("source")} }

// Email returns an OR text filter for the email field.
func (m *orModifier) Email() TextFilter { return TextFilter{m.base("email")} }

// Phone returns an OR text filter for the phone field.
func (m *orModifier) Phone() TextFilter { return TextFilter{m.base("phone")} }

// Website returns an OR text filter for the website field.
func (m *orModifier) Website() TextFilter { return TextFilter{m.base("website")} }

// Rating returns an OR number filter for the rating field.
func (m *orModifier) Rating() NumberFilter { return NumberFilter{m.base("rating")} }

// ReviewCount returns an OR number filter for the review_count field.
func (m *orModifier) ReviewCount() NumberFilter { return NumberFilter{m.base("review_count")} }

// Tags returns an OR array filter for the tags field.
func (m *orModifier) Tags() ArrayFilter { return ArrayFilter{m.base("tags")} }

// Location returns an OR location filter.
func (m *orModifier) Location() LocationFilter { return LocationFilter{m.base("location")} }

// Attr returns an OR attribute filter for the named custom attribute.
func (m *orModifier) Attr(name string) AttrFilter {
	return AttrFilter{m.base("attr:" + name)}
}

// --- Top-level AND filter constructors ---

func andBase(field string) filterBase {
	return filterBase{logic: logicAnd, field: field}
}

// City returns an AND text filter for the city field.
func City() TextFilter { return TextFilter{andBase("city")} }

// Name returns an AND text filter for the name field.
func Name() TextFilter { return TextFilter{andBase("name")} }

// Country returns an AND text filter for the country field.
func Country() TextFilter { return TextFilter{andBase("country")} }

// State returns an AND text filter for the state field.
func State() TextFilter { return TextFilter{andBase("state")} }

// Category returns an AND text filter for the category field.
func Category() TextFilter { return TextFilter{andBase("category")} }

// Source returns an AND text filter for the source field.
func Source() TextFilter { return TextFilter{andBase("source")} }

// Email returns an AND text filter for the email field.
func Email() TextFilter { return TextFilter{andBase("email")} }

// Phone returns an AND text filter for the phone field.
func Phone() TextFilter { return TextFilter{andBase("phone")} }

// Website returns an AND text filter for the website field.
func Website() TextFilter { return TextFilter{andBase("website")} }

// Rating returns an AND number filter for the rating field.
func Rating() NumberFilter { return NumberFilter{andBase("rating")} }

// ReviewCount returns an AND number filter for the review_count field.
func ReviewCount() NumberFilter { return NumberFilter{andBase("review_count")} }

// Tags returns an AND array filter for the tags field.
func Tags() ArrayFilter { return ArrayFilter{andBase("tags")} }

// Location returns an AND location filter.
func Location() LocationFilter { return LocationFilter{andBase("location")} }

// Attr returns an AND attribute filter for the named custom attribute.
// The field is serialized as "attr:<name>" in the filter expression.
//
// Example:
//
//	klozeo.Attr("industry").Eq("Software")
func Attr(name string) AttrFilter {
	return AttrFilter{andBase("attr:" + name)}
}

