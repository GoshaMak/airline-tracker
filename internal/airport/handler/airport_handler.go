package handler

import (
	"airline-tracker/internal/airport/command"
	"airline-tracker/internal/airport/dto"
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

	g := r.Group("/airport", middleware.AuthMiddleware("admin"))
	{
		g.POST("/create", c.CreateAirport)
		g.POST("/list", c.ListAirports)
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
	req := &dto.CreateAirportRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	cmd, err := command.NewCreateAirportCommand(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	if err := h.uc.CreateAirport(cmd); err != nil {
		if errors.Is(err, usecase.ErrAirportAlreadyExists) {
			ctx.JSON(http.StatusConflict, gin.H{"msg": "airport already exists"})
			return
		}
		slog.Info("handler.create_airport", "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"msg": "airport created"})
}

func (h *AirportHandler) ListAirports(ctx *gin.Context) {

}
