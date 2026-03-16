package controller

import (
	"airline-tracker/internal/controller/dto"
	"airline-tracker/internal/middleware"
	"airline-tracker/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type AdminController struct {
	service *service.AdminService
}

func NewAdminController(i do.Injector) (*AdminController, error) {
	return &AdminController{
		service: do.MustInvoke[*service.AdminService](i),
	}, nil
}

func RegisterAdminRoutes(i do.Injector, r *gin.Engine) {
	c := do.MustInvoke[*AdminController](i)
	g := r.Group("/admin", middleware.AuthMiddleware("admin"))
	{
		g.POST("/flight", c.AddFlight)
		g.PATCH("/flight/:id", c.UpdateFlight)
		g.DELETE("/flight/:id", c.DeleteFlight)

	}

	{
		g.POST("/aircraft_model", c.AddAircraftModel)
		g.POST("/aircraft", c.AddAircraft)
		g.POST("/airport", c.AddAirport)
		g.POST("/gate", c.AddGate)
	}
}

func (c *AdminController) AddFlight(ctx *gin.Context) {
	type Req struct {
		Flight           dto.FlightDTO   `json:"flight"`
		Aircraft         dto.AircraftDTO `json:"aircraft"`
		DepartureAirport dto.AirportDTO  `json:"departure_airport"`
		ArrivalAirport   dto.AirportDTO  `json:"arrival_airport"`
		DepartureGate    dto.GateDTO     `json:"departure_gate"`
		ArrivalGate      dto.GateDTO     `json:"arrival_gate"`
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

func (c *AdminController) UpdateFlight(ctx *gin.Context) {}

func (c *AdminController) DeleteFlight(ctx *gin.Context) {}

func (c *AdminController) AddAircraft(ctx *gin.Context) {
	type Req struct {
		Aircraft dto.AircraftDTO `json:"aircraft"`
	}
	req := &Req{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse args"})
		return
	}
	if err := c.service.AddAircraft(req.Aircraft.AircraftFromDTO()); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "aircraft created"})
}

func (c *AdminController) AddAirport(ctx *gin.Context) {
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

func (c *AdminController) AddGate(ctx *gin.Context) {
	type Req struct {
		Gate dto.GateDTO `json:"gate"`
	}
	req := &Req{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse args"})
		return
	}
	if err := c.service.AddGate(req.Gate.GateFromDTO()); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "gate created"})
}

func (c *AdminController) AddAircraftModel(ctx *gin.Context) {
	type Req struct {
		AircraftModel dto.AircraftModelDTO `json:"aircraft_model"`
	}
	req := &Req{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse args"})
		return
	}
	if err := c.service.AddAircraftModel(req.AircraftModel.AircraftModelFromDTO()); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "aircraft model created"})
}
