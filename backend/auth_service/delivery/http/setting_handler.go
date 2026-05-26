package http

import (
	"net/http"
	"backend/domain"
	"backend/delivery/http/middleware"
	"github.com/gin-gonic/gin"
)

type SettingHandler struct {
	settingUsecase domain.SettingUsecase
}

func NewSettingHandler(r *gin.RouterGroup, u domain.SettingUsecase) {
	handler := &SettingHandler{settingUsecase: u}
	
	settings := r.Group("/settings")
	settings.Use(middleware.JWTAuthMiddleware())
	{
		settings.GET("", handler.GetSettings)
		settings.PUT("", handler.UpdateSettings)
	}
}

type UpdateSettingRequest struct {
	ErpType           string `json:"erp_type" binding:"required"`
	Endpoint          string `json:"endpoint" binding:"required,url"`
	ApiKey            string `json:"api_key" binding:"required"`
	VolumeSensitivity int    `json:"volume_sensitivity" binding:"min=0,max=100"`
	GeoThreshold      int    `json:"geo_threshold" binding:"min=0,max=100"`
	VelocityLimit     int    `json:"velocity_limit" binding:"min=0,max=100"`
}

func (h *SettingHandler) GetSettings(c *gin.Context) {
	userId := c.MustGet("userId").(uint)

	config, err := h.settingUsecase.GetSettings(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to retrieve settings",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Settings retrieved successfully",
		"data":    config,
	})
}

func (h *SettingHandler) UpdateSettings(c *gin.Context) {
	userId := c.MustGet("userId").(uint)

	var req UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid request payload",
			"error":   err.Error(),
		})
		return
	}

	config := &domain.Configuration{
		ErpType:           req.ErpType,
		Endpoint:          req.Endpoint,
		ApiKey:            req.ApiKey,
		VolumeSensitivity: req.VolumeSensitivity,
		GeoThreshold:      req.GeoThreshold,
		VelocityLimit:     req.VelocityLimit,
	}

	updated, err := h.settingUsecase.UpdateSettings(c.Request.Context(), userId, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to update settings",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Settings updated successfully",
		"data":    updated,
	})
}
