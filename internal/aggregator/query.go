package aggregator

type SortField string

const (
	SortByTitle     SortField = "title"
	SortByUpdatedAt SortField = "updated_at"
)

type SortOrder string

const (
	OrderAsc  SortOrder = "asc"
	OrderDesc SortOrder = "desc"
)

type Query struct {
	Sort  SortField
	Order SortOrder
	Page  int //start from 1
	Size  int
}

// Result is one page of merged documents plus each source's status.
type Result struct {
	Documents []Document
	Total     int // total documents across all pages
	Page      int
	HasNext   bool
	Statuses  []SourceStatus
}
