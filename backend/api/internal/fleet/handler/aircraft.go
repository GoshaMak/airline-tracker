package handler

import (
	"api/internal/fleet/command"
	"api/internal/fleet/dto"
	"api/internal/fleet/usecase"
	"api/internal/httperror"
	"api/internal/middleware"
	"api/internal/pagination"
	userDomain "api/internal/user/domain"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type AircraftHandler struct {
	uc *usecase.AircraftUsecase
}

func NewAircraftHandler(i do.Injector) (*AircraftHandler, error) {
	return &AircraftHandler{
		uc: do.MustInvoke[*usecase.AircraftUsecase](i),
	}, nil
}

func RegisterAircraftRoutes(i do.Injector, r *gin.RouterGroup) {
	h := do.MustInvoke[*AircraftHandler](i)

	admin := r.Group("/aircraft", middleware.AuthMiddleware(userDomain.AdminRole))
	{
		admin.POST("", h.CreateAircraft)
		admin.GET("", h.ListAircrafts)
	}
}

func (h *AircraftHandler) CreateAircraft(ctx *gin.Context) {
	const op = "AircraftHandler.AddAircraft"
	req := &dto.CreateAircraftRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse args"})
		return
	}
	cmd, err := command.NewCreateAircraftCommand(req)
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	if err := h.uc.CreateAircraft(cmd); err != nil {
		if errors.Is(err, usecase.ErrAircraftAlreadyExists) {
			ctx.JSON(http.StatusCreated, gin.H{"msg": "aircraft created"})
			return
		}
		slog.Error(op, "err", err)
		httperror.Internal(ctx, err, usecase.ErrOperationNotPermitted)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"msg": "aircraft created"})
}

func (h *AircraftHandler) ListAircrafts(ctx *gin.Context) {
	const op = "AircraftHandler.ListAircrafts"
	params, err := pagination.Parse(ctx.Query("limit"), ctx.Query("cursor"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	page, err := h.uc.ListAircraftsPage(params)
	if err != nil {
		slog.Error(op, "err", err)
		httperror.Internal(ctx, err, usecase.ErrOperationNotPermitted)
		return
	}
	resp := dto.ToResponseListAircrafts(page)
	ctx.JSON(http.StatusOK, resp)
}
