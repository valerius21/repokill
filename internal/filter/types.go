// Package filter provides types for repository filtering and sorting.
package filter

// SortField defines the field to sort by.
type SortField int

const (
	SortByPushedAt SortField = iota
	SortByName
	SortByStars
)

// SortOrder defines the direction of sorting.
type SortOrder int

const (
	Ascending SortOrder = iota
	Descending
)

// SortOptions combines field and order for sorting.
type SortOptions struct {
	Field SortField
	Order SortOrder
}

// FilterOptions defines criteria for filtering repositories.
type FilterOptions struct {
	Archived    *bool
	Forked      *bool
	Visibility  string
	SearchQuery string
}
