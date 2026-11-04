package httpapi

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mtuan8820/unified-document-viewer/api"
	"github.com/mtuan8820/unified-document-viewer/internal/aggregator"
)

// Defaults and limits from api/openapi.yaml.
const (
	defaultPage = 1
	defaultSize = 20
	maxSize     = 100
)

type Handler struct {
	agg *aggregator.Aggregator
}

func NewHandler(agg *aggregator.Aggregator) *Handler { return &Handler{agg: agg} }

var _ api.ServerInterface = (*Handler)(nil)

func (h *Handler) FetchVehicleDocuments(c *gin.Context, vin string, params api.FetchVehicleDocumentsParams) {
	query, err := queryFromParams(params)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.Error{Error: err.Error()})
		return
	}

	result, err := h.agg.GetDocuments(c.Request.Context(), vin, query)
	if errors.Is(err, aggregator.ErrAllSourcesFailed) {
		c.JSON(http.StatusBadGateway, api.Error{Error: "all document sources are unavailable"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toResponse(result))
}

// toResponse maps the aggregator's Result to the generated API type.
func toResponse(result aggregator.Result) api.DocumentListResponse {
	// make, not nil, so an empty page is encoded as [] rather than null.
	docs := make([]api.Document, 0, len(result.Documents))
	for _, d := range result.Documents {
		docs = append(docs, api.Document{
			DocumentId: d.ID,
			Source:     api.Source(d.Source),
			Title:      d.Title,
			Type:       d.Type,
			UpdatedAt:  d.UpdatedAt,
		})
	}

	statuses := make([]api.SourceStatus, 0, len(result.Statuses))
	for _, s := range result.Statuses {
		statuses = append(statuses, api.SourceStatus{
			Source:    api.Source(s.Source),
			Status:    api.SourceStatusStatus(s.Status),
			FetchedAt: s.FetchedAt, // nil for timeout/error, omitted in JSON
		})
	}

	return api.DocumentListResponse{
		Data:     docs,
		TotalDoc: result.Total,
		Page:     result.Page,
		HasNext:  result.HasNext,
		Status:   statuses,
	}
}

// queryFromParams applies the API defaults and validates the values,
// because the generated wrapper only checks types, not enums or ranges.
func queryFromParams(params api.FetchVehicleDocumentsParams) (aggregator.Query, error) {
	query := aggregator.Query{
		Sort:  aggregator.SortByTitle,
		Order: aggregator.OrderAsc,
		Page:  defaultPage,
		Size:  defaultSize,
	}

	if params.Sort != nil {
		if !params.Sort.Valid() {
			return aggregator.Query{}, fmt.Errorf("invalid sort %q: must be title or updated_at", *params.Sort)
		}
		query.Sort = aggregator.SortField(*params.Sort)
	}

	if params.Order != nil {
		if !params.Order.Valid() {
			return aggregator.Query{}, fmt.Errorf("invalid order %q: must be asc or desc", *params.Order)
		}
		query.Order = aggregator.SortOrder(*params.Order)
	}

	if params.Page != nil {
		if *params.Page < 1 {
			return aggregator.Query{}, fmt.Errorf("invalid page %d: must be >= 1", *params.Page)
		}
		query.Page = *params.Page
	}

	if params.Size != nil {
		if *params.Size < 1 || *params.Size > maxSize {
			return aggregator.Query{}, fmt.Errorf("invalid size %d: must be between 1 and %d", *params.Size, maxSize)
		}
		query.Size = *params.Size
	}

	return query, nil
}
