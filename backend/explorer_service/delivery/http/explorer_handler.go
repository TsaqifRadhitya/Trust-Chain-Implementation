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
		explorerRoutes.GET("/blockchain/validate", handler.ValidateChain)
<<<<<<< HEAD
		explorerRoutes.GET("/stats", handler.GetStats)
=======
		explorerRoutes.POST("/transactions/:hash/correct", handler.AddCorrection)
>>>>>>> e857de251da799f90dbbe5428e7c365c73a9e567
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

func (h *ExplorerHandler) ValidateChain(c *gin.Context) {
	result, err := h.usecase.ValidateChain()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "message": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "message": "Blockchain validation completed", "data": result})
}

<<<<<<< HEAD
func (h *ExplorerHandler) GetStats(c *gin.Context) {
	stats, err := h.usecase.GetStats()
=======
type CorrectionRequest struct {
	ActualStatus string `json:"actual_status" binding:"required"`
	Reason       string `json:"reason"`
	CorrectedBy  string `json:"corrected_by"`
}

func (h *ExplorerHandler) AddCorrection(c *gin.Context) {
	hash := c.Param("hash")
	var req CorrectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "message": err.Error(), "data": nil})
		return
	}

	correction, err := h.usecase.AddCorrection(hash, req.ActualStatus, req.Reason, req.CorrectedBy)
>>>>>>> e857de251da799f90dbbe5428e7c365c73a9e567
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "message": err.Error(), "data": nil})
		return
	}
<<<<<<< HEAD
	c.JSON(http.StatusOK, gin.H{"status": 200, "message": "Dashboard stats retrieved", "data": stats})
}
=======

	c.JSON(http.StatusOK, gin.H{"status": 200, "message": "Correction registered successfully", "data": correction})
}

>>>>>>> e857de251da799f90dbbe5428e7c365c73a9e567
