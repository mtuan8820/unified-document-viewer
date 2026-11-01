package aggregator

import "time"

// SourceName identifies an external system. Together with Document.ID it
// uniquely identifies a document, because IDs can collide across systems.
type SourceName string

const (
	SourceSales   SourceName = "sales"
	SourceService SourceName = "service"
)

// Document is normalized document metadata. Adapters produce it; the
// snapshot and cache store it as JSON.
type Document struct {
	ID        string     `json:"document_id"`
	Source    SourceName `json:"source"`
	Title     string     `json:"title"`
	Type      string     `json:"type"`
	UpdatedAt time.Time  `json:"updated_at"` // when the document last changed upstream
}
