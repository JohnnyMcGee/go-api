package db

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetHandler(c *gin.Context) {
	c.JSON(http.StatusOK, "cowabunga")
}
