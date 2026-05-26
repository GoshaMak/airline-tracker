package handler

import (
	"api/internal/flight/dto"
	"api/internal/middleware"
	"api/internal/user/domain"
	"api/internal/user/usecase"
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

func RegisterRoutes(i do.Injector, r *gin.Engine) {
	h := do.MustInvoke[*UserHandler](i)

	user := r.Group("/user", middleware.AuthMiddleware(domain.UserRole))
	{
		user.POST("/subscribe", h.Subscribe)
		user.GET("/list_flights", h.ListFlights)
	}
}

// пришла подписка -> записать в подписки -> записать в outbox событие для кафки
// тоже самое если обновление полей полёта -> записать в outbox событие для кафки
//
// publisher читает неотправленные в outbox -> отправляет в кафку сообщение -> отмечает отправленным (в outbox'е)
//
// notifier получает сообщение из кафки -> отправляет сообщение об изменении/о вылете
// notifier:
// если подписка, то записывает в свою таблицу из которой читает worker и отправляет, когда надо
// если обновление, то обновляет таблицу и отправляет сообщение об изменении

// @Summary subscribe user (only user)
// @Tags User
// @Security BearerAuth
// @Param flight_id query string true "flight id"
// @Produce json
// @Success 200 "user subscribed"
// @Failure 401
// @Failure 404
// @Failure 500
// @Router /user/subscribe [post]
func (h *UserHandler) Subscribe(ctx *gin.Context) {
	const op = "UserHandler.Subscribe"
	uidStr := ctx.GetString("user_id")
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		slog.Error(op, "err", err)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"msg": "unauthorized",
		})
		return
	}

	fidStr := ctx.Query("flight_id")
	fid, err := uuid.Parse(fidStr)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"msg": "flight not found",
		})
		return
	}

	slog.Debug(op, "uid", uid, "fid", fid)

	if err := h.uc.Subscribe(uid, fid); err != nil {
		switch err {
		case usecase.ErrUserNotFound:
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"msg": "user not found",
			})
			return
		case usecase.ErrFlightNotFound:
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"msg": "flight not found",
			})
			return
		}
		slog.Warn(op, "err", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"msg": "internal error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"msg": "user subscribed",
	})
}

// @Summary list flights (only user)
// @Description get all flights in which user is subscribed
// @Tags User
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200
// @Failure 400
// @Failure 401
// @Router /user/list_flights [get]
func (h *UserHandler) ListFlights(ctx *gin.Context) {
	const op = "UserHandler.ListFlights"
	uidStr := ctx.GetString("user_id")
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		slog.Error(op, "err", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"msg": "unauthorized",
		})
		return
	}

	flights, err := h.uc.ListFlights(uid)
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": "internal error",
		})
		return
	}

	resp, err := dto.ToResponseListFlights(flights)
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": "internal error",
		})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
