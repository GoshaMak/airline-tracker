package handler

import (
	"airline-tracker/internal/airport/command"
	"airline-tracker/internal/airport/dto"
	"airline-tracker/internal/airport/usecase"
	"airline-tracker/internal/common"
	"airline-tracker/internal/middleware"
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

func RegisterGateRoutes(i do.Injector, r *gin.Engine) {
	h := do.MustInvoke[*GateHandler](i)

	admin := r.Group("/gate", middleware.AuthMiddleware(common.AdminRole))
	{
		admin.POST("/create", h.CreateGate)
		admin.GET("/list", h.ListGates)
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
// @Failure 409
// @Failure 500
// @Router /gate/create [post]
func (h *GateHandler) CreateGate(ctx *gin.Context) {
	req := &dto.CreateGateRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	cmd, err := command.NewCreateGateCommand(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	if err := h.uc.CreateGate(cmd); err != nil {
		if errors.Is(err, usecase.ErrGateAlreadyExists) {
			ctx.JSON(http.StatusConflict, gin.H{"msg": "gate already exists"})
			return
		}
		slog.Info("handler.create_gate", "err", err)
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
// @Success 200 ""
// @Failure 400
// @Failure 500
// @Router /gate/list [get]
func (h *GateHandler) ListGates(ctx *gin.Context) {}
