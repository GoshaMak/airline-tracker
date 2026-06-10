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

func RegisterAircraftRoutes(i do.Injector, r *gin.Engine) {
	h := do.MustInvoke[*AircraftHandler](i)

	admin := r.Group("/aircraft", middleware.AuthMiddleware(userDomain.AdminRole))
	{
		admin.POST("/create", h.CreateAircraft)
		admin.GET("/list", h.ListAircrafts)
	}
}

// @Summary add aircraft (only admin)
// @Description create new aircraft
// @Tags Aircraft
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param aircraft body dto.CreateAircraftRequest true "aircraft info"
// @Success 201 "created"
// @Failure 400
// @Failure 401
// @Failure 500
// @Router /aircraft/create [post]
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"msg": "aircraft created"})
}

// @Summary list aircrafts (only admin)
// @Description list aircrafts
// @Tags Aircraft
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {array} dto.ListAircraftsResponse
// @Failure 401
// @Failure 500
// @Router /aircraft/list [get]
func (h *AircraftHandler) ListAircrafts(ctx *gin.Context) {
	const op = "AircraftHandler.ListAircrafts"
	as, err := h.uc.ListAircrafts()
	if err != nil {
		slog.Error(op, "err", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}
	resp := dto.ToResponseListAircrafts(as)
	ctx.JSON(http.StatusOK, resp)
}
