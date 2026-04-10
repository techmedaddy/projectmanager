package db

const (
	// DefaultPage is the starting page used when pagination is not specified.
	DefaultPage = 1
	// DefaultLimit is the default number of rows returned per page.
	DefaultLimit = 20
	// MaxLimit prevents list endpoints from requesting unbounded result sizes.
	MaxLimit = 100
)

// Pagination provides shared page/limit normalization for list endpoints.
type Pagination struct {
	Page  int
	Limit int
}

// Normalize applies defaults and enforces sensible bounds.
func (p Pagination) Normalize() Pagination {
	if p.Page < 1 {
		p.Page = DefaultPage
	}

	if p.Limit < 1 {
		p.Limit = DefaultLimit
	}

	if p.Limit > MaxLimit {
		p.Limit = MaxLimit
	}

	return p
}

// Offset returns the SQL offset for the current page and limit.
func (p Pagination) Offset() int {
	normalized := p.Normalize()
	return (normalized.Page - 1) * normalized.Limit
}
