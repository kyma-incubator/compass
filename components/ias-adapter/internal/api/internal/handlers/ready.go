package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReadyHandler struct {
}

func (h ReadyHandler) Ready(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
}
