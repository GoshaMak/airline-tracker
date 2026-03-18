package controller

import (
	"airline-tracker/internal/flight/dto"
	"airline-tracker/internal/flight/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type FlightController struct {
	service *service.FlightService
}

func NewFlightController(i do.Injector) (*FlightController, error) {
	return &FlightController{
		service: do.MustInvoke[*service.FlightService](i),
	}, nil
}

func RegisterFlightRoutes(i do.Injector, r *gin.Engine) {
	c := do.MustInvoke[*FlightController](i)
	r.GET("/flights", c.Flights)
	r.GET("/flight/:id", c.FlightByID)
}

func (c *FlightController) Flights(ctx *gin.Context) {
	flights, err := c.service.ListAllFlights()
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "error"})
		return
	}
	flights_dto := make([]dto.FlightDTO, len(flights), cap(flights))
	for i, fl := range flights {
		flights_dto[i] = *dto.FlightToDTO(&fl)
	}
	ctx.JSON(http.StatusOK, gin.H{"flights": flights_dto})
}

func (c *FlightController) FlightByID(ctx *gin.Context) {}
