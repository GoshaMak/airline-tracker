package controller

import (
	"airline-tracker/internal/flight/command"
	"airline-tracker/internal/flight/dto"
	"airline-tracker/internal/flight/service"
	"airline-tracker/internal/middleware"
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

func RegisterRoutes(i do.Injector, r *gin.Engine) {
	c := do.MustInvoke[*FlightController](i)

	{
		r.GET("/flights", c.Flights)
		r.GET("/flight/:id", c.FlightByID)
	}

	g := r.Group("/admin", middleware.AuthMiddleware("admin"))
	{
		g.POST("/add_flight", c.AddFlight)
		g.PATCH("/flight/:id", c.UpdateFlight) // TODO: move id to request's body
		g.DELETE("/flight/:id", c.DeleteFlight)
	}
}

// @Summary list all flights
// @Description list all flights
// @Tags flight
// @Accept json
// @Produce json
// @Success 200 {array} dto.FlightDTO
// @Failure 400
// @Router /flights [get]
func (c *FlightController) Flights(ctx *gin.Context) {
	flights, err := c.service.ListAllFlights()
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "error"})
		return
	}
	flights_dto := make([]dto.FlightDTO, len(flights), cap(flights))
	ctx.JSON(http.StatusOK, gin.H{"flights": flights_dto})
}

func (c *FlightController) FlightByID(ctx *gin.Context) {}

// @Summary add flight
// @Description create a flight
// @Tags flight
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param flight body dto.CreateFlightRequest true "flight info"
// @Success 201
// @Failure 400
// @Failure 500
// @Router /admin/add_flight [post]
func (c *FlightController) AddFlight(ctx *gin.Context) {
	req := &dto.CreateFlightRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse args", "err": err})
		return
	}
	cmd, err := command.NewAddFlightCommand(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request", "err": err})
		return
	}
	if err := c.service.AddFlight(cmd); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"msg": "flight created"})
}

func (c *FlightController) UpdateFlight(ctx *gin.Context) {}

func (c *FlightController) DeleteFlight(ctx *gin.Context) {}
