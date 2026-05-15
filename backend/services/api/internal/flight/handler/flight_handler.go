package handler

import (
	"api/internal/flight/command"
	"api/internal/flight/dto"
	"api/internal/flight/usecase"
	"api/internal/middleware"
	userDomain "api/internal/user/domain"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type FlightHandler struct {
	uc *usecase.FlightUsecase
}

func NewFlightHandler(i do.Injector) (*FlightHandler, error) {
	return &FlightHandler{
		uc: do.MustInvoke[*usecase.FlightUsecase](i),
	}, nil
}

func RegisterRoutes(i do.Injector, r *gin.Engine) {
	c := do.MustInvoke[*FlightHandler](i)

	{
		r.GET("/list_flights", c.ListFlights)
		r.GET("/flight/:id", c.FlightById)
	}

	admin := r.Group("", middleware.AuthMiddleware(userDomain.AdminRole))
	{
		admin.POST("/add_flight", c.CreateFlight)
		admin.PATCH("/flight/:id", c.UpdateFlight)
		admin.DELETE("/flight/:id", c.DeleteFlight)
	}
}

// @Summary list all flights
// @Description list all flights
// @Tags Flight
// @Accept json
// @Produce json
// @Success 200 {array} dto.ListFlightsResponse
// @Failure 400
// @Failure 500
// @Router /list_flights [get]
func (h *FlightHandler) ListFlights(ctx *gin.Context) {
	const op = "FlightHandler.ListFlights"
	flights, err := h.uc.ListFlights()
	if err != nil && err != usecase.ErrCacheSave {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": "internal error",
		})
		return
	}

	response, err := dto.ToResponseListFlights(flights)
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": "internal error",
		})
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (h *FlightHandler) FlightById(ctx *gin.Context) {}

// @Summary create flight (only admin)
// @Description create a flight
// @Tags Flight
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param flight body dto.CreateFlightRequest true "flight info"
// @Success 201
// @Failure 400
// @Failure 500
// @Router /create_flight [post]
func (h *FlightHandler) CreateFlight(ctx *gin.Context) {
	req := &dto.CreateFlightRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse args", "err": err})
		return
	}
	cmd, err := command.NewCreateFlightCommand(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request", "err": err})
		return
	}
	if err := h.uc.CreateFlight(cmd); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"msg": "flight created"})
}

func (h *FlightHandler) UpdateFlight(ctx *gin.Context) {}

func (h *FlightHandler) DeleteFlight(ctx *gin.Context) {}
