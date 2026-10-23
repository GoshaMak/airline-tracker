package handler

import (
	"api/internal/flight/command"
	flightDTO "api/internal/flight/dto"
	"api/internal/flight/usecase"
	"api/internal/httperror"
	"api/internal/middleware"
	"api/internal/pagination"
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

func RegisterRoutes(i do.Injector, r *gin.RouterGroup) {
	c := do.MustInvoke[*FlightHandler](i)

	{
		r.GET("/flights", c.ListFlights)
		r.GET("/flights/:id", c.FlightById)
	}

	admin := r.Group("/flights", middleware.AuthMiddleware(userDomain.AdminRole))
	{
		admin.POST("", c.CreateFlight)
		admin.PATCH("/:id", c.UpdateFlight)
	}
}

func (h *FlightHandler) ListFlights(ctx *gin.Context) {
	const op = "FlightHandler.ListFlights"
	params, err := pagination.Parse(ctx.Query("limit"), ctx.Query("cursor"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	page, err := h.uc.ListFlightsPage(ctx.Request.Context(), params)
	if err != nil {
		slog.ErrorContext(ctx.Request.Context(), op, "err", err)
		httperror.Internal(ctx, err, usecase.ErrOperationNotPermitted)
		return
	}

	response := flightDTO.ToResponseListFlights(page)
	ctx.JSON(http.StatusOK, response)
}

func (h *FlightHandler) FlightById(ctx *gin.Context) {
	const op = "FlightHandler.FlightById"
	fidStr := ctx.Param("id")
	fid, err := uuid.Parse(fidStr)
	if err != nil {
		slog.WarnContext(ctx.Request.Context(), op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	fd, err := h.uc.FlightById(ctx.Request.Context(), fid)
	if err != nil {
		if errors.Is(err, usecase.ErrFlightNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"msg": "flight not found"})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), op, "err", err)
		httperror.Internal(ctx, err, usecase.ErrOperationNotPermitted)
		return
	}

	fi := flightDTO.ToFlightInfoDomain(&fd)
	resp := flightDTO.ListFlightByIdResponse{Flight: fi}
	ctx.JSON(http.StatusOK, resp)
}

func (h *FlightHandler) CreateFlight(ctx *gin.Context) {
	const op = "FlightHandler.CreateFlight"
	req := &flightDTO.CreateFlightRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		slog.WarnContext(ctx.Request.Context(), op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	cmd, err := command.NewCreateFlightCommand(req)
	if err != nil {
		slog.WarnContext(ctx.Request.Context(), op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	if err := h.uc.CreateFlight(ctx.Request.Context(), cmd); err != nil { // FIX: returns 500 if flight exists
		slog.WarnContext(ctx.Request.Context(), op, "err", err)
		httperror.Internal(ctx, err, usecase.ErrOperationNotPermitted)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"msg": "flight created"})
}

func (h *FlightHandler) UpdateFlight(ctx *gin.Context) {
	const op = "FlightHandler.UpdateFlight"
	req := &flightDTO.UpdateFlightRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		slog.WarnContext(ctx.Request.Context(), op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		slog.WarnContext(ctx.Request.Context(), op, "err", err)
		ctx.JSON(http.StatusNotFound, gin.H{"msg": "not found"})
		return
	}
	req.Flight.FlightId = id

	cmd, err := command.NewUpdateFlightCommand(req)
	if err != nil {
		slog.WarnContext(ctx.Request.Context(), op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request", "err": err})
		return
	}
	if err := h.uc.UpdateFlight(ctx.Request.Context(), cmd); err != nil {
		if errors.Is(err, usecase.ErrFlightNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"msg": "flight not found"})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), op, "err", err)
		httperror.Internal(ctx, err, usecase.ErrOperationNotPermitted)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "flight updated"})
}
