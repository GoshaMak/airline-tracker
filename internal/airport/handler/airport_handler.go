package handler

import (
	"airline-tracker/internal/airport/command"
	"airline-tracker/internal/airport/dto"
	"airline-tracker/internal/airport/usecase"
	"airline-tracker/internal/middleware"
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
	g := r.Group("/admin", middleware.AuthMiddleware("admin"))
	{
		g.POST("/add_airport", c.AddAirport)
	}
}

// @Summary new airport
// @Description creates an airport
// @Tags airport
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param airport body dto.CreateAirportRequest true "airport info"
// @Success 201 "airport created"
// @Failure 400
// @Failure 500
// @Router /admin/add_airport [post]
func (h *AirportHandler) AddAirport(ctx *gin.Context) {
	req := &dto.CreateAirportRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse args"})
		return
	}
	cmd, err := command.NewAddAirportCommand(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "false args"})
		return
	}
	if err := h.uc.AddAirport(cmd); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "bad request"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "airport created"})
}
