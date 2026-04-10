package handler

import (
	"airline-tracker/internal/flight/command"
	"airline-tracker/internal/flight/dto"
	"airline-tracker/internal/flight/usecase"
	"airline-tracker/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type FlightHandler struct {
	uc *service.FlightUsecase
}

func NewFlightHandler(i do.Injector) (*FlightHandler, error) {
	return &FlightHandler{
		uc: do.MustInvoke[*service.FlightUsecase](i),
	}, nil
}

func RegisterRoutes(i do.Injector, r *gin.Engine) {
	c := do.MustInvoke[*FlightHandler](i)

	{
		r.GET("/list_flights", c.ListFlights)
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
// @Success 200 {array} dto.ListFlightsResponse
// @Failure 400
// @Router /list_flights [get]
func (h *FlightHandler) ListFlights(ctx *gin.Context) {
	flights, err := h.uc.ListAllFlights()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"err": err})
		return
	}
	response, err := dto.ToResponseListFlights(flights)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"flights": *response})
}

func (h *FlightHandler) FlightByID(ctx *gin.Context) {}

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
func (h *FlightHandler) AddFlight(ctx *gin.Context) {
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
	if err := h.uc.AddFlight(cmd); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"msg": "flight created"})
}

func (h *FlightHandler) UpdateFlight(ctx *gin.Context) {}

func (h *FlightHandler) DeleteFlight(ctx *gin.Context) {}
