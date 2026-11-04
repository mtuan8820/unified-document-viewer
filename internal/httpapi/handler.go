package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mtuan8820/unified-document-viewer/api"
	"github.com/mtuan8820/unified-document-viewer/internal/aggregator"
)

type Handler struct {
	agg *aggregator.Aggregator
}

func NewHandler(agg *aggregator.Aggregator) *Handler { return &Handler{agg: agg} }

var _ api.ServerInterface = (*Handler)(nil)

func (h *Handler) FetchVehicleDocuments(c *gin.Context, vin string, params api.FetchVehicleDocumentsParams) {
	docs, err := h.agg.GetDocuments(c.Request.Context(), vin)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": docs})
}
