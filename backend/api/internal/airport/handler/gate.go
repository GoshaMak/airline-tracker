package handler

import (
	"api/internal/airport/command"
	"api/internal/airport/dto"
	"api/internal/airport/usecase"
	"api/internal/middleware"
	"api/internal/pagination"
	userDomain "api/internal/user/domain"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type GateHandler struct {
	uc *usecase.GateUsecase
}

func NewGateHandler(i do.Injector) (*GateHandler, error) {
	return &GateHandler{
		uc: do.MustInvoke[*usecase.GateUsecase](i),
	}, nil
}

func RegisterGateRoutes(i do.Injector, r *gin.RouterGroup) {
	h := do.MustInvoke[*GateHandler](i)

	admin := r.Group("/gates", middleware.AuthMiddleware(userDomain.AdminRole))
	{
		admin.POST("", h.CreateGate)
		admin.GET("", h.ListGates)
	}
}

// @Summary create gate (only admin)
// @Tags Gate
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param gate body dto.CreateGateRequest true "gate info"
// @Success 201 "gate created"
// @Failure 400
// @Failure 401
// @Failure 500
// @Router /api/v1/gates [post]
func (h *GateHandler) CreateGate(ctx *gin.Context) {
	const op = "GateHandler.CreateGate"
	req := &dto.CreateGateRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	cmd, err := command.NewCreateGateCommand(req)
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	if err := h.uc.CreateGate(cmd); err != nil {
		if errors.Is(err, usecase.ErrGateAlreadyExists) {
			ctx.JSON(http.StatusCreated, gin.H{"msg": "gate created"})
			return
		}
		slog.Error(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"msg": "gate created"})
}

// @Summary list gates (only admin)
// @Tags Gate
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} dto.ListGatesResponse
// @Failure 401
// @Failure 500
// @Router /api/v1/gates [get]
func (h *GateHandler) ListGates(ctx *gin.Context) {
	const op = "GateHandler.ListGates"
	params, err := pagination.Parse(ctx.Query("limit"), ctx.Query("cursor"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	page, err := h.uc.ListGatesPage(params)
	if err != nil {
		slog.Error(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}

	resp := dto.ToResponseListGates(page)
	ctx.JSON(http.StatusOK, resp)
}
