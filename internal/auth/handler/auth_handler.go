package handler

import (
	"airline-tracker/internal/auth/dto"
	"airline-tracker/internal/auth/usecase"
	userDomain "airline-tracker/internal/user/domain"
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
	c := do.MustInvoke[*AuthHandler](i)

	r.POST("register", c.Register)
	r.POST("login", c.Login)
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
	var req dto.CreateUserDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	u := userDomain.NewUser(req.Email, req.Password, req.Role)
	if err := h.uc.CreateUser(u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
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
	var req dto.LoginRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}

	user, err := h.uc.GetUser(req.Email, req.Password)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"msg": "user not found"})
		return
	}

	token, err := usecase.GenerateJWT(user)
	if err != nil {
		slog.Warn("Error generating jwt", "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "err"})
		return
	}

	ctx.JSON(http.StatusOK, dto.LoginResponseDTO{Token: token})
}
