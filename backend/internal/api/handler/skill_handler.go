package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/insmtx/Leros/backend/internal/api/contract"
	"github.com/insmtx/Leros/backend/internal/api/dto"
)

// RegisterSkillRoutes registers skill management routes.
func RegisterSkillRoutes(r gin.IRouter, service contract.SkillService) {
	r.PATCH("/skills/:code/status", toggleSkillStatus(service))
}

func toggleSkillStatus(service contract.SkillService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		code := strings.TrimSpace(ctx.Param("code"))
		if code == "" {
			ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, "code is required"))
			return
		}

		var req contract.ToggleSkillStatusRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
			return
		}

		result, err := service.ToggleSkillStatus(ctx, code, &req)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				ctx.JSON(http.StatusNotFound, dto.Error(dto.CodeNotFound, err.Error()))
				return
			}
			if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "required") {
				ctx.JSON(http.StatusBadRequest, dto.Error(dto.CodeInvalidParams, err.Error()))
				return
			}
			ctx.JSON(http.StatusInternalServerError, dto.Error(dto.CodeInternalError, err.Error()))
			return
		}

		ctx.JSON(http.StatusOK, dto.Success(result))
	}
}
