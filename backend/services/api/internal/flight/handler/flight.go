package handler

import (
	"api/internal/flight/command"
	"api/internal/flight/dto"
	"api/internal/flight/usecase"
	"api/internal/middleware"
	userDomain "api/internal/user/domain"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		r.GET("/flight/list", c.ListFlights)
		r.GET("/flight/:id", c.FlightById)
	}

	admin := r.Group("/flight", middleware.AuthMiddleware(userDomain.AdminRole))
	{
		admin.POST("/create", c.CreateFlight)
		admin.PATCH("/:id", c.UpdateFlight)
	}
}

// @Summary list all flights
// @Description list all flights
// @Tags Flight
// @Accept json
// @Produce json
// @Success 200 {array} dto.ListFlightsResponse
// @Failure 500
// @Router /flight/list [get]
func (h *FlightHandler) ListFlights(ctx *gin.Context) {
	const op = "FlightHandler.ListFlights"
	flights, err := h.uc.ListFlights()
	if err != nil {
		slog.Error(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}

	response := dto.ToResponseListFlights(flights)
	ctx.JSON(http.StatusOK, response)
}

// @Summary flight info
// @Description get flight info by its id
// @Tags Flight
// @Accept json
// @Param id path string true "flight id"
// @Produce json
// @Success 200 {object} dto.ListFlightByIdResponse
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /flight/{id} [get]
func (h *FlightHandler) FlightById(ctx *gin.Context) {
	const op = "FlightHandler.FlightById"
	fidStr := ctx.Param("id")
	fid, err := uuid.Parse(fidStr)
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	fd, err := h.uc.FlightById(fid)
	if err != nil {
		if errors.Is(err, usecase.ErrFlightNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"mgs": "flight not found"})
			return
		}
		slog.Error(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}

	fi := dto.ToFlightInfoDomain(&fd)
	resp := dto.ListFlightByIdResponse{Flight: fi}
	ctx.JSON(http.StatusOK, resp)
}

// @Summary create flight (only admin)
// @Description create a flight
// @Tags Flight
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param flight body dto.CreateFlightRequest true "flight info"
// @Success 201 "flight created"
// @Failure 400
// @Failure 401
// @Failure 500
// @Router /flight/create [post]
func (h *FlightHandler) CreateFlight(ctx *gin.Context) {
	const op = "FlightHandler.CreateFlight"
	req := &dto.CreateFlightRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	cmd, err := command.NewCreateFlightCommand(req)
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	if err := h.uc.CreateFlight(cmd); err != nil { // FIX: returns 500 if flight exists
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"msg": "flight created"})
}

// @Summary update flight (only admin)
// @Description updates flight. able to update: all deps, status, plan
// @Tags Flight
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "flight id"
// @Param flight body dto.UpdateFlightRequest true "flight info"
// @Success 200
// @Failure 400
// @Failure 401
// @Failure 404
// @Failure 500
// @Router /flight/{id} [patch]
func (h *FlightHandler) UpdateFlight(ctx *gin.Context) {
	const op = "FlightHandler.UpdateFlight"
	req := &dto.UpdateFlightRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusNotFound, gin.H{"msg": "not found"})
		return
	}
	req.Flight.FlightId = id

	cmd, err := command.NewUpdateFlightCommand(req)
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request", "err": err})
		return
	}
	if err := h.uc.UpdateFlight(cmd); err != nil {
		if errors.Is(err, usecase.ErrFlightNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"msg": "flight not found"})
			return
		}
		slog.Error(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "flight updated"})
}
