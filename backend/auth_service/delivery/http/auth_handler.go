package http

import (
	"fmt"
	"net/http"
	"strconv"
	"backend/domain"
	"backend/delivery/http/middleware"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUsecase domain.AuthUsecase
}

func NewAuthHandler(r *gin.RouterGroup, u domain.AuthUsecase) {
	handler := &AuthHandler{authUsecase: u}
	
	r.POST("/auth/login", handler.Login)
	r.POST("/auth/refresh", handler.Refresh)
	r.GET("/auth/validate", middleware.JWTOptionalMiddleware(), handler.Validate)
	
	// Internal routes (no auth middleware)
	r.GET("/internal/users", handler.GetAllUsers)
	r.GET("/internal/users/:id", handler.GetUserByID)
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid request payload",
			"error":   err.Error(),
		})
		return
	}

	accessToken, refreshToken, user, err := h.authUsecase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  http.StatusUnauthorized,
			"message": "Login failed",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Login successful",
		"data": gin.H{
			"token":         accessToken,
			"refresh_token": refreshToken,
			"user": gin.H{
				"email": user.Email,
				"name":  user.Name,
			},
		},
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid request payload",
			"error":   err.Error(),
		})
		return
	}

	newAccessToken, newRefreshToken, err := h.authUsecase.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  http.StatusUnauthorized,
			"message": "Token refresh failed",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Token refreshed successfully",
		"data": gin.H{
			"token":         newAccessToken,
			"refresh_token": newRefreshToken,
		},
	})
}

// Validate is an endpoint for Nginx auth_request module.
// It relies on the JWTAuthMiddleware to validate the token.
// If the token is valid, it returns a 200 OK along with the X-User-Id header.
func (h *AuthHandler) Validate(c *gin.Context) {
	userId, exists := c.Get("userId")
	if exists {
		// Pass user ID back so Nginx can forward it to upstream
		c.Header("X-User-Id", fmt.Sprintf("%v", userId))
	}
	c.Status(http.StatusOK)
}

// GetUserByID is an internal route, should be called by other services, not exposed externally via Nginx
func (h *AuthHandler) GetUserByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid user ID",
			"error":   err.Error(),
		})
		return
	}

	user, err := h.authUsecase.GetUserByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  http.StatusNotFound,
			"message": "User not found",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "User found",
		"data": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
	})
}

// GetAllUsers is an internal route to get all users
func (h *AuthHandler) GetAllUsers(c *gin.Context) {
	users, err := h.authUsecase.GetAllUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to fetch users",
			"error":   err.Error(),
		})
		return
	}

	var responseData []gin.H
	for _, user := range users {
		responseData = append(responseData, gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		})
	}

	// If no users, return empty array instead of null
	if responseData == nil {
		responseData = []gin.H{}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Users found",
		"data":    responseData,
	})
}
