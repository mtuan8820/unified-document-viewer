package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mtuan8820/unified-document-viewer/internal/aggregator"
)

type ServiceSystemAdapter struct {
	baseURL string
	client  *http.Client
}

type serviceResponse struct {
	Data []serviceRecord `json:"data"`
}

type serviceRecord struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	DocType   string    `json:"doc_type"`
	UpdatedAt time.Time `json:"updated_at"`
}

var _ aggregator.Source = (*ServiceSystemAdapter)(nil)

func NewServiceSystemAdapter(baseURL string, client *http.Client) *ServiceSystemAdapter {
	return &ServiceSystemAdapter{baseURL: baseURL, client: client}
}

func (ex *ServiceSystemAdapter) FetchDocuments(ctx context.Context, vin string) ([]aggregator.Document, error) {
	u := ex.baseURL + "/service/records?" + url.Values{"vin": {vin}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("service: build request: %w", err)
	}

	resp, err := ex.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("service: call: %w", err)
	}
	defer resp.Body.Close()

	// The service system returns 204 when the VIN has no records: no documents, not a failure.
	if resp.StatusCode == http.StatusNoContent {
		return []aggregator.Document{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("service: unexpected status %d", resp.StatusCode)
	}

	var body serviceResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("service: decode body: %w", err)
	}

	docs := make([]aggregator.Document, 0, len(body.Data))
	for _, r := range body.Data {
		docs = append(docs, serviceRecordToDocument(r))
	}
	return docs, nil
}

func (ex *ServiceSystemAdapter) Name() aggregator.SourceName {
	return aggregator.SourceService
}

func serviceRecordToDocument(r serviceRecord) aggregator.Document {
	return aggregator.Document{
		ID:        r.ID,
		Source:    aggregator.SourceService,
		Title:     r.Title,
		Type:      strings.ToLower(r.DocType),
		UpdatedAt: r.UpdatedAt,
	}
}
