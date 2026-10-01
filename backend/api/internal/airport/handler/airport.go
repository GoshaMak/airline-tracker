package handler

import (
	"api/internal/airport/command"
	"api/internal/airport/dto"
	"api/internal/airport/query"
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

type AirportHandler struct {
	uc *usecase.AirportUsecase
}

func NewAirportHandler(i do.Injector) (*AirportHandler, error) {
	return &AirportHandler{
		uc: do.MustInvoke[*usecase.AirportUsecase](i),
	}, nil
}

func RegisterAirportRoutes(i do.Injector, r *gin.RouterGroup) {
	h := do.MustInvoke[*AirportHandler](i)

	admin := r.Group("/airports", middleware.AuthMiddleware(userDomain.AdminRole))
	{
		admin.POST("", h.CreateAirport)
	}

	all := r.Group("/airports")
	{
		all.GET("", h.ListAirports)
	}
}

// @Summary create airport (only admin)
// @Tags Airport
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param airport body dto.CreateAirportRequest true "airport info"
// @Success 201 "airport created"
// @Failure 400
// @Failure 401
// @Failure 500
// @Router /api/v1/airports [post]
func (h *AirportHandler) CreateAirport(ctx *gin.Context) {
	const op = "AirportHandler.CreateAirport"
	req := &dto.CreateAirportRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	cmd, err := command.NewCreateAirportCommand(req)
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	if err := h.uc.CreateAirport(cmd); err != nil {
		if errors.Is(err, usecase.ErrAirportAlreadyExists) {
			ctx.JSON(http.StatusOK, gin.H{"msg": "airport created"})
			return
		}
		slog.Error(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"msg": "airport created"})
}

// @Summary list airports
// @Tags Airport
// @Produce json
// @Success 200 {object} dto.ListAirportsResponse
// @Failure 500
// @Router /api/v1/airports [get]
func (h *AirportHandler) ListAirports(ctx *gin.Context) {
	const op = "AirportHandler.ListAirports"
	params, err := pagination.Parse(ctx.Query("limit"), ctx.Query("cursor"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	page, err := h.uc.ListAirportsPage(params)
	if err != nil {
		slog.Error(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}
	resp := query.QueryToListAirportsResponse(page)
	ctx.JSON(http.StatusOK, resp)
}
