package controller

import (
	airportDTO "airline-tracker/internal/airport/dto"
	fleetDTO "airline-tracker/internal/fleet/dto"
	"airline-tracker/internal/flight/dto"
	flightDTO "airline-tracker/internal/flight/dto"
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
		g.POST("/flight", c.AddFlight)
		g.PATCH("/flight/:id", c.UpdateFlight) // TODO: move id to request's body
		g.DELETE("/flight/:id", c.DeleteFlight)
	}
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

func (c *FlightController) AddFlight(ctx *gin.Context) {
	type Req struct {
		Flight           flightDTO.FlightDTO   `json:"flight"`
		Aircraft         fleetDTO.AircraftDTO  `json:"aircraft"`
		DepartureAirport airportDTO.AirportDTO `json:"departure_airport"`
		ArrivalAirport   airportDTO.AirportDTO `json:"arrival_airport"`
		DepartureGate    airportDTO.GateDTO    `json:"departure_gate"`
		ArrivalGate      airportDTO.GateDTO    `json:"arrival_gate"`
	}
	req := &Req{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse args"})
		return
	}
	if err := c.service.AddFlight(
		req.Flight.FlightFromDTO(), req.Aircraft.AircraftFromDTO(),
		req.DepartureAirport.AirportFromDTO(), req.ArrivalAirport.AirportFromDTO(),
		req.DepartureGate.GateFromDTO(), req.ArrivalGate.GateFromDTO()); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"msg": "flight created"})
}

func (c *FlightController) UpdateFlight(ctx *gin.Context) {}

func (c *FlightController) DeleteFlight(ctx *gin.Context) {}
