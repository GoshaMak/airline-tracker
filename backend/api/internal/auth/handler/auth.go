package handler

import (
	"api/internal/auth/dto"
	"api/internal/auth/usecase"
	"api/internal/httperror"
	"api/internal/middleware"
	userDomain "api/internal/user/domain"
	"errors"
	"log/slog"
	"net/http"
	"shared/common"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	r.POST("/auth/verify", h.VerifyLogin)
	r.POST("/auth/recovery", h.RequestRecovery)
	r.POST("/auth/recovery/verify", h.VerifyRecovery)
	r.POST("/auth/password", middleware.AuthAnyRoleMiddleware(), h.ChangePassword)
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
	email, err := common.NewEmail(req.Email)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	challengeID, err := h.uc.BeginLogin(ctx.Request.Context(), email, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidPassword) || errors.Is(err, usecase.ErrAccountLocked) {
			status := http.StatusUnauthorized
			msg := "invalid credentials"
			if errors.Is(err, usecase.ErrAccountLocked) {
				status, msg = http.StatusLocked, "account locked"
			}
			ctx.JSON(status, gin.H{"msg": msg})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), op, "err", err)
		httperror.Internal(ctx, err, usecase.ErrOperationNotPermitted)
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{"challenge_id": challengeID})
}

func (h *AuthHandler) VerifyLogin(ctx *gin.Context) {
	var req struct {
		ChallengeID string `json:"challenge_id" binding:"required,uuid"`
		Code        string `json:"code" binding:"required,len=6,numeric"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	token, err := h.uc.VerifyLogin(ctx.Request.Context(), req.ChallengeID, req.Code)
	if err != nil {
		if errors.Is(err, usecase.ErrAccountLocked) {
			ctx.JSON(http.StatusLocked, gin.H{"msg": "account locked"})
			return
		}
		if errors.Is(err, usecase.ErrInvalidCode) || errors.Is(err, usecase.ErrInvalidPassword) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "invalid or expired code"})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), "AuthHandler.VerifyLogin", "err", err)
		httperror.Internal(ctx, err, usecase.ErrOperationNotPermitted)
		return
	}
	ctx.JSON(http.StatusOK, dto.LoginResponseDTO{Token: token})
}

func (h *AuthHandler) RequestRecovery(ctx *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	email, err := common.NewEmail(req.Email)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	if err := h.uc.RequestRecovery(ctx.Request.Context(), email); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "AuthHandler.RequestRecovery", "err", err)
		httperror.Internal(ctx, err, usecase.ErrOperationNotPermitted)
		return
	}
	ctx.JSON(http.StatusAccepted, gin.H{"msg": "if the account is locked, a recovery code was sent"})
}

func (h *AuthHandler) VerifyRecovery(ctx *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	email, err := common.NewEmail(req.Email)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	if err := h.uc.VerifyRecovery(ctx.Request.Context(), email, req.Code); err != nil {
		if errors.Is(err, usecase.ErrInvalidCode) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "invalid or expired code"})
			return
		}
		httperror.Internal(ctx, err, usecase.ErrOperationNotPermitted)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "account unlocked"})
}

func (h *AuthHandler) ChangePassword(ctx *gin.Context) {
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	uid, err := uuid.Parse(ctx.GetString("user_id"))
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "unauthorized"})
		return
	}
	if err := h.uc.ChangePassword(ctx.Request.Context(), uid, req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, usecase.ErrInvalidPassword) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "invalid credentials"})
			return
		}
		if errors.Is(err, usecase.ErrWeakPassword) {
			ctx.JSON(http.StatusBadRequest, gin.H{"msg": "new password does not meet policy"})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), "AuthHandler.ChangePassword", "err", err)
		httperror.Internal(ctx, err, usecase.ErrOperationNotPermitted)
		return
	}
	ctx.Status(http.StatusNoContent)
}
