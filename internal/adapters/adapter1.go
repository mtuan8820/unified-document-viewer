package adapters

import (
	"context"

	"github.com/mtuan8820/unified-document-viewer/internal/aggregator"
)

type ExternalServiceAdapterOne struct {
}

var _ aggregator.Source = (*ExternalServiceAdapterOne)(nil)

func (ex *ExternalServiceAdapterOne) FetchDocuments(ctx context.Context, vin string) ([]aggregator.Document, error) {
	return nil, nil
}
