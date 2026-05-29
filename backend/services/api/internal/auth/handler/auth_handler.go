package handler

import (
	"api/internal/auth/dto"
	"api/internal/auth/usecase"
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

func RegisterAuthRoutes(i do.Injector, r *gin.Engine) {
	h := do.MustInvoke[*AuthHandler](i)

	r.POST("register", h.Register)
	r.POST("login", h.Login)
}

// @Summary register
// @Description creates a user
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body dto.CreateUserDTO true "user info"
// @Success 201 "user created"
// @Failure 400
// @Failure 500
// @Router /register [post]
func (h *AuthHandler) Register(ctx *gin.Context) {
	const op = "AuthHandler.Register"
	var req dto.CreateUserDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	u, err := userDomain.NewUser(req.Email, req.Password, req.Role)
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "bad request",
		})
		return
	}
	if err := h.uc.CreateUser(u); err != nil {
		if errors.Is(err, usecase.ErrUserAlreadyExists) {
			slog.Warn(op, "err", err)
			ctx.JSON(http.StatusConflict, gin.H{
				"msg": "email is used",
			})
			return
		}
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": "internal error",
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"msg": "user created"})
}

// @Summary login
// @Description user authentication
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body dto.LoginRequestDTO true "user info"
// @Success 200 {object} dto.LoginResponseDTO "auth token"
// @Failure 401
// @Failure 404
// @Failure 500
// @Router /login [post]
func (h *AuthHandler) Login(ctx *gin.Context) {
	const op = "AuthHandler.Login"
	var req dto.LoginRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	slog.Debug(op, "req", req)

	user, err := h.uc.GetUser(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			slog.Warn(op, "err", err)
			ctx.JSON(http.StatusNotFound, gin.H{"msg": "user not found"})
			return
		}
		if errors.Is(err, usecase.ErrWrongPassword) {
			slog.Warn(op, "err", err)
			ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "wrong password"})
			return
		}
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}
	slog.Info(op, "user", user)

	token, err := usecase.GenerateJWT(user)
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err})
		return
	}

	ctx.JSON(http.StatusOK, dto.LoginResponseDTO{Token: token})
}
