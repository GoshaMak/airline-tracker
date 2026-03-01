package controller

import (
	"airline-tracker/internal/service"

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
	r.POST("/flights", c.Flights)
	r.GET("/flights/:id", c.FlightsByID)
}

func (c *FlightController) Flights(ctx *gin.Context) {

}

func (c *FlightController) FlightsByID(ctx *gin.Context)
