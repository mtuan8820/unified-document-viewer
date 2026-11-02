package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mtuan8820/unified-document-viewer/internal/aggregator"
)

type SalesSystemAdapter struct {
	baseURL string
	client  *http.Client
}

type saleDocument struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Category     string    `json:"category"`
	LastModified time.Time `json:"lastModified"`
}

type salesRepsonse struct {
	Documents []saleDocument `json:"documents"`
}

var _ aggregator.Source = (*SalesSystemAdapter)(nil)

func NewSalesSystemAdapter(baseURL string, client *http.Client) *SalesSystemAdapter {
	return &SalesSystemAdapter{baseURL: baseURL, client: client}
}

func (ex *SalesSystemAdapter) FetchDocuments(ctx context.Context, vin string) ([]aggregator.Document, error) {
	url := ex.baseURL + "/api/v1/vehicles/" + vin + "/documents"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("sales: build request: %w", err)
	}

	resp, err := ex.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sales: call: %w", err)
	}
	defer resp.Body.Close()

	// The sales system returns 404 for an unknown VIN: no documents, not a failure.
	if resp.StatusCode == http.StatusNotFound {
		return []aggregator.Document{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sales: unexpected status %d", resp.StatusCode)
	}

	var body salesRepsonse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("sales: decode body: %w", err)
	}

	docs := make([]aggregator.Document, 0, len(body.Documents))
	for _, r := range body.Documents {
		docs = append(docs, saleDocumentToDocument(r))
	}
	return docs, nil

}

func (ex *SalesSystemAdapter) Name() aggregator.SourceName {
	return aggregator.SourceSales
}

func saleDocumentToDocument(s saleDocument) aggregator.Document {
	return aggregator.Document{
		ID:        s.ID,
		Source:    aggregator.SourceSales,
		Title:     s.Name,
		Type:      strings.ToLower(s.Category),
		UpdatedAt: s.LastModified,
	}
}
