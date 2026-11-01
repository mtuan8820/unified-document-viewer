package aggregator

import "context"

type Source interface {
	FetchDocuments(ctx context.Context, vin string) ([]Document, error)
}
