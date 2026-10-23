package handler

import (
	flightDTO "api/internal/flight/dto"
	"api/internal/httperror"
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

func (h *UserHandler) Subscribe(ctx *gin.Context) {
	const op = "UserHandler.Subscribe"
	uidStr := ctx.GetString("user_id")
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		slog.WarnContext(ctx.Request.Context(), op, "err", err)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"msg": "user not found"})
		return
	}

	var req userDTO.CreateSubscriptionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		slog.WarnContext(ctx.Request.Context(), op, "err", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	if err := h.uc.CreateSubscription(ctx.Request.Context(), uid, req.FlightID, *req.NotifyBeforeMinutes); err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"msg": "user not found"})
			return
		}
		if errors.Is(err, usecase.ErrFlightNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"msg": "flight not found"})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), op, "err", err)
		httperror.Internal(ctx, err, usecase.ErrOperationNotPermitted)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "user subscribed"})
}

func (h *UserHandler) AddTimers(ctx *gin.Context) {
	const op = "UserHandler.AddTimers"
	uid, err := uuid.Parse(ctx.GetString("user_id"))
	if err != nil {
		slog.WarnContext(ctx.Request.Context(), op, "err", err)
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
		slog.WarnContext(ctx.Request.Context(), op, "err", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	if err := h.uc.AddSubscriptionTimers(ctx.Request.Context(), uid, fid, req.NotifyBeforeMinutes); err != nil {
		if errors.Is(err, usecase.ErrSubscriptionNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"msg": "subscription not found"})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), op, "err", err)
		httperror.Internal(ctx, err, usecase.ErrOperationNotPermitted)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "notification timers added"})
}

func (h *UserHandler) Unsubscribe(ctx *gin.Context) {
	const op = "UserHandler.Unsubscribe"
	uid, err := uuid.Parse(ctx.GetString("user_id"))
	if err != nil {
		slog.WarnContext(ctx.Request.Context(), op, "err", err)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"msg": "user not found"})
		return
	}
	fid, err := uuid.Parse(ctx.Param("flight_id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	if err := h.uc.DeleteSubscription(ctx.Request.Context(), uid, fid); err != nil {
		slog.ErrorContext(ctx.Request.Context(), op, "err", err)
		httperror.Internal(ctx, err, usecase.ErrOperationNotPermitted)
		return
	}
	ctx.Status(http.StatusNoContent)
}

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
		slog.ErrorContext(ctx.Request.Context(), op, "err", err)
		ctx.JSON(http.StatusNotFound, gin.H{"msg": "user not found"})
		return
	}

	page, err := h.uc.ListFlightsPage(ctx.Request.Context(), uid, params)
	if err != nil {
		slog.ErrorContext(ctx.Request.Context(), op, "err", err)
		httperror.Internal(ctx, err, usecase.ErrOperationNotPermitted)
		return
	}

	resp := flightDTO.ToResponseListFlights(page)
	ctx.JSON(http.StatusOK, resp)
}
