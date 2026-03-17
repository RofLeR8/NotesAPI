package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /api/health - проверить статуст приложения
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
