package handler

import (
	"api/internal/fleet/command"
	"api/internal/fleet/dto"
	"api/internal/fleet/usecase"
	"api/internal/middleware"
	userDomain "api/internal/user/domain"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type AircraftModelHandler struct {
	uc *usecase.AircraftModelUsecase
}

func NewAircraftModelHandler(i do.Injector) (*AircraftModelHandler, error) {
	return &AircraftModelHandler{
		uc: do.MustInvoke[*usecase.AircraftModelUsecase](i),
	}, nil
}

func RegisterAircraftModelRoutes(i do.Injector, r *gin.RouterGroup) {
	h := do.MustInvoke[*AircraftModelHandler](i)

	admin := r.Group("/aircraft-models", middleware.AuthMiddleware(userDomain.AdminRole))
	{
		admin.POST("", h.CreateAircraftModel)
		admin.GET("/:id", h.AircraftModelById)
	}
}

func (h *AircraftModelHandler) CreateAircraftModel(ctx *gin.Context) {
	const op = "AircraftModelHandler.AddAircraftModel"
	req := &dto.CreateAircraftModelRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	cmd, err := command.NewCreateAircraftModelCommand(req)
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	if err := h.uc.CreateAircraftModel(cmd); err != nil {
		if errors.Is(err, usecase.ErrAircraftModelAlreadyExists) {
			ctx.JSON(http.StatusCreated, gin.H{"msg": "aircraft model created"})
			return
		}
		slog.Error(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"msg": "aircraft model created"})
}

func (h *AircraftModelHandler) AircraftModelById(ctx *gin.Context) {
	const op = "AircraftModelHandler.AircraftModelById"
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusNotFound, gin.H{"msg": "not found"})
		return
	}

	am, err := h.uc.AircraftById(id)
	if err != nil {
		if errors.Is(err, usecase.ErrAircraftModelNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"msg": "not found"})
		}
		slog.Error(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}

	resp := dto.ToResponseAircraftModelInfo(am)
	ctx.JSON(http.StatusOK, resp)
}
