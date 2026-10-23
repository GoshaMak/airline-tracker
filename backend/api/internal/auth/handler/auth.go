package handler

import (
	"api/internal/auth/dto"
	"api/internal/auth/usecase"
	"api/internal/httperror"
	userDomain "api/internal/user/domain"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type AuthHandler struct {
	uc *usecase.AuthUsecase
}

func NewAuthHandler(i do.Injector) (*AuthHandler, error) {
	return &AuthHandler{
		uc: do.MustInvoke[*usecase.AuthUsecase](i),
	}, nil
}

func RegisterAuthRoutes(i do.Injector, r *gin.RouterGroup) {
	h := do.MustInvoke[*AuthHandler](i)

	r.POST("/users", h.Register)
	r.POST("/auth/tokens", h.Login)
}

func (h *AuthHandler) Register(ctx *gin.Context) {
	const op = "AuthHandler.Register"
	var req dto.CreateUserDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		slog.WarnContext(ctx.Request.Context(), op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	u, err := userDomain.NewUser(req.Email, req.Password, userDomain.UserRole)
	if err != nil {
		slog.WarnContext(ctx.Request.Context(), op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	if err := h.uc.CreateUser(ctx.Request.Context(), u); err != nil {
		if errors.Is(err, usecase.ErrUserAlreadyExists) {
			ctx.JSON(http.StatusConflict, gin.H{"msg": "email is used"})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), op, "err", err)
		httperror.Internal(ctx, err, usecase.ErrOperationNotPermitted)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"msg": "user created"})
}

func (h *AuthHandler) Login(ctx *gin.Context) {
	const op = "AuthHandler.Login"
	var req dto.LoginRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		slog.WarnContext(ctx.Request.Context(), op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	user, err := h.uc.GetUser(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"msg": "user not found"})
			return
		}
		if errors.Is(err, usecase.ErrInvalidPassword) {
			ctx.JSON(http.StatusNotFound, gin.H{"msg": "user not found"})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), op, "err", err)
		httperror.Internal(ctx, err, usecase.ErrOperationNotPermitted)
		return
	}

	token, err := usecase.GenerateJWT(user)
	if err != nil {
		slog.ErrorContext(ctx.Request.Context(), op, "err", err)
		httperror.Internal(ctx, err, usecase.ErrOperationNotPermitted)
		return
	}

	ctx.JSON(http.StatusOK, dto.LoginResponseDTO{Token: token})
}
