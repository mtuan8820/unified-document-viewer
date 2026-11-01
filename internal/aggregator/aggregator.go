package aggregator

import (
	"context"

	"github.com/go-redis/cache/v9"
	"golang.org/x/sync/errgroup"
)

type Aggregator struct {
	cache   *cache.Cache
	sources []Source
}

func (a *Aggregator) GetDocuments(ctx context.Context, vin string) ([]Document, error) {
	var group errgroup.Group
	ret := make([]Document, 0)

	for _, source := range a.sources {
		group.Go(func() error {
			docs, err := source.FetchDocuments(ctx, vin)
			if err != nil {
				//error handling, print status or do something
			} else {
				ret = append(ret, docs...)
			}

			return nil
		})
	}

	err := group.Wait()
	if err != nil {
		return nil, err
	}
	return ret, nil
}
