package http

import (
	"explorer_service/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ExplorerHandler struct {
	usecase usecase.ExplorerUsecase
}

func NewExplorerHandler(r *gin.RouterGroup, u usecase.ExplorerUsecase) {
	handler := &ExplorerHandler{usecase: u}

	explorerRoutes := r.Group("/explorer")
	{
		explorerRoutes.GET("/blocks", handler.GetRecentBlocks)
		explorerRoutes.GET("/blocks/:id", handler.GetBlockDetail)
		explorerRoutes.GET("/transactions", handler.GetRecentTransactions)
		explorerRoutes.GET("/transactions/:hash", handler.GetTransactionDetail)
		explorerRoutes.GET("/address/:address", handler.GetAddressDetail)
		explorerRoutes.GET("/search", handler.Search)
	}
}

func (h *ExplorerHandler) GetRecentBlocks(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	blocks, err := h.usecase.GetRecentBlocks(limit, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": 200, "message": "Recent blocks retrieved successfully", "data": blocks})
}

func (h *ExplorerHandler) GetBlockDetail(c *gin.Context) {
	id := c.Param("id")
	block, err := h.usecase.GetBlockDetail(id)
	if err != nil || block == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": 404, "message": "Block not found", "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "message": "Block details retrieved", "data": block})
}

func (h *ExplorerHandler) GetRecentTransactions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	txs, err := h.usecase.GetRecentTransactions(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": 200, "message": "Recent transactions retrieved", "data": txs})
}

func (h *ExplorerHandler) GetTransactionDetail(c *gin.Context) {
	hash := c.Param("hash")
	tx, err := h.usecase.GetTransactionDetail(hash)
	if err != nil || tx == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": 404, "message": "Transaction not found", "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "message": "Transaction details retrieved", "data": tx})
}

func (h *ExplorerHandler) GetAddressDetail(c *gin.Context) {
	address := c.Param("address")
	data, err := h.usecase.GetAddressDetail(address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "message": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "message": "Address details retrieved", "data": data})
}

func (h *ExplorerHandler) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "message": "Query param 'q' is required", "data": nil})
		return
	}

	result, err := h.usecase.Search(q)
	if err != nil || result == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": 404, "message": "Search result not found", "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "message": "Search result found", "data": result})
}
