package handler

import (
	flightDTO "api/internal/flight/dto"
	"api/internal/middleware"
	"api/internal/pagination"
	"api/internal/user/domain"
	userDTO "api/internal/user/dto"
	"api/internal/user/usecase"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type UserHandler struct {
	uc *usecase.UserUsecase
}

func NewUserHandler(i do.Injector) (*UserHandler, error) {
	return &UserHandler{
		uc: do.MustInvoke[*usecase.UserUsecase](i),
	}, nil
}

func RegisterRoutes(i do.Injector, r *gin.RouterGroup) {
	h := do.MustInvoke[*UserHandler](i)

	subscriptions := r.Group("/subscriptions", middleware.AuthMiddleware(domain.UserRole))
	{
		subscriptions.POST("", h.Subscribe)
		subscriptions.GET("", h.ListFlights)
		subscriptions.DELETE("/:flight_id", h.Unsubscribe)
		subscriptions.POST("/:flight_id/timers", h.AddTimers)
	}
}

// @Summary subscribe user (only user)
// @Tags User
// @Security BearerAuth
// @Param subscription body userDTO.CreateSubscriptionRequest true "subscription"
// @Produce json
// @Success 200 "user subscribed"
// @Failure 401
// @Failure 404
// @Failure 500
// @Router /api/v1/subscriptions [post]
func (h *UserHandler) Subscribe(ctx *gin.Context) {
	const op = "UserHandler.Subscribe"
	uidStr := ctx.GetString("user_id")
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"msg": "user not found"})
		return
	}

	var req userDTO.CreateSubscriptionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		slog.Warn(op, "err", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	if err := h.uc.CreateSubscription(uid, req.FlightID, *req.NotifyBeforeMinutes); err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"msg": "user not found"})
			return
		}
		if errors.Is(err, usecase.ErrFlightNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"msg": "flight not found"})
			return
		}
		slog.Error(op, "err", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "user subscribed"})
}

func (h *UserHandler) AddTimers(ctx *gin.Context) {
	const op = "UserHandler.AddTimers"
	uid, err := uuid.Parse(ctx.GetString("user_id"))
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"msg": "user not found"})
		return
	}
	fid, err := uuid.Parse(ctx.Param("flight_id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	var req userDTO.AddSubscriptionTimersRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		slog.Warn(op, "err", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	if err := h.uc.AddSubscriptionTimers(uid, fid, req.NotifyBeforeMinutes); err != nil {
		if errors.Is(err, usecase.ErrSubscriptionNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"msg": "subscription not found"})
			return
		}
		slog.Error(op, "err", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "notification timers added"})
}

func (h *UserHandler) Unsubscribe(ctx *gin.Context) {
	const op = "UserHandler.Unsubscribe"
	uid, err := uuid.Parse(ctx.GetString("user_id"))
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"msg": "user not found"})
		return
	}
	fid, err := uuid.Parse(ctx.Param("flight_id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	if err := h.uc.DeleteSubscription(uid, fid); err != nil {
		slog.Error(op, "err", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}
	ctx.Status(http.StatusNoContent)
}

// @Summary list flights (only user)
// @Description get all flights in which user is subscribed
// @Tags User
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of items to return" default(50) minimum(1) maximum(100)
// @Param cursor query string false "Opaque cursor returned as next_cursor by the previous page" minlength(1)
// @Success 200 {object} flightDTO.ListFlightsResponse
// @Failure 400
// @Failure 401
// @Failure 500
// @Router /api/v1/subscriptions [get]
func (h *UserHandler) ListFlights(ctx *gin.Context) {
	const op = "UserHandler.ListFlights"
	params, err := pagination.Parse(ctx.Query("limit"), ctx.Query("cursor"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	uidStr := ctx.GetString("user_id")
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		slog.Error(op, "err", err)
		ctx.JSON(http.StatusNotFound, gin.H{"msg": "user not found"})
		return
	}

	page, err := h.uc.ListFlightsPage(uid, params)
	if err != nil {
		slog.Error(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}

	resp := flightDTO.ToResponseListFlights(page)
	ctx.JSON(http.StatusOK, resp)
}
