package controller

import (
	"airline-tracker/internal/airport/dto"
	"airline-tracker/internal/airport/service"
	"airline-tracker/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type AirportController struct {
	service *service.AirportService
}

func NewAirportController(i do.Injector) (*AirportController, error) {
	return &AirportController{
		service: do.MustInvoke[*service.AirportService](i),
	}, nil
}

func RegisterAirportRoutes(i do.Injector, r *gin.Engine) {
	c := do.MustInvoke[*AirportController](i)
	g := r.Group("/admin", middleware.AuthMiddleware("admin"))
	{
		g.POST("/airport", c.AddAirport)
	}
}

func (c *AirportController) AddAirport(ctx *gin.Context) {
	type Req struct {
		Airport dto.AirportDTO `json:"airport"`
	}
	req := &Req{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse args"})
		return
	}
	if err := c.service.AddAirport(req.Airport.AirportFromDTO()); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "airport created"})
}
