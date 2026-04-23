package handler

import (
	"airline-tracker/internal/airport/command"
	"airline-tracker/internal/airport/dto"
	"airline-tracker/internal/airport/query"
	"airline-tracker/internal/airport/usecase"
	"airline-tracker/internal/middleware"
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

func RegisterAirportRoutes(i do.Injector, r *gin.Engine) {
	c := do.MustInvoke[*AirportHandler](i)

	admin := r.Group("/airport", middleware.AuthMiddleware("admin"))
	{
		admin.POST("/create", c.CreateAirport)
	}

	user := r.Group("/airport")
	{
		user.GET("/list", c.ListAirports)
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
// @Failure 409
// @Failure 500
// @Router /airport/create [post]
func (h *AirportHandler) CreateAirport(ctx *gin.Context) {
	op := "AirportHandler.CreateAirport"
	req := &dto.CreateAirportRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		slog.Debug(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	cmd, err := command.NewCreateAirportCommand(req)
	if err != nil {
		slog.Debug(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	if err := h.uc.CreateAirport(cmd); err != nil {
		if errors.Is(err, usecase.ErrAirportAlreadyExists) {
			ctx.JSON(http.StatusConflict, gin.H{"msg": "airport already exists"})
			return
		}
		slog.Info(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"msg": "airport created"})
}

// @Summary list airports
// @Tags Airport
// @Produce json
// @Success 200 {array} dto.ListAirportsResponse
// @Failure 500
// @Router /airport/list [get]
func (h *AirportHandler) ListAirports(ctx *gin.Context) {
	op := "AirportHandler.ListAirports"
	airports, err := h.uc.ListAirports()
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": "internal error",
		})
		return
	}

	resp := query.QueryToListAirportsResponse(airports)
	ctx.JSON(http.StatusOK, resp)
}
