package controller

import (
	"airline-tracker/internal/auth/dto"
	"airline-tracker/internal/auth/service"
	userDomain "airline-tracker/internal/user/domain"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type AuthController struct {
	service *service.AuthService
}

func NewAuthController(i do.Injector) (*AuthController, error) {
	return &AuthController{
		service: do.MustInvoke[*service.AuthService](i),
	}, nil
}

func RegisterAuthRoutes(i do.Injector, r *gin.Engine) {
	c := do.MustInvoke[*AuthController](i)

	g := r.Group("/auth")
	{
		g.POST("signup", c.Signup)
		g.POST("login", c.Login)
	}
}

// @Summary signup
// @Description creates a user
// @Tags auth
// @Accept json
// @Produce json
// @Param user body dto.SignupDTO true "user info"
// @Success 201 "user created"
// @Failure 400
// @Failure 500
// @Router /auth/signup [post]
func (c *AuthController) Signup(ctx *gin.Context) {
	var req dto.SignupDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse args"})
		return
	}
	u := userDomain.NewUser(req.Email, req.Phone, req.Password, req.Role)
	if err := c.service.CreateUser(u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "failed to create a user"})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"msg": "user created"})
}

// @Summary login
// @Description user authentication
// @Tags auth
// @Accept json
// @Produce json
// @Param user body dto.LoginRequestDTO true "user info"
// @Success 200 {object} dto.LoginResponseDTO "auth token"
// @Failure 401
// @Failure 500
// @Router /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req dto.LoginRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "incorrect args"})
		return
	}
	slog.Debug("Auth", "email", req.Email, "phone", req.Phone, "pswd", req.Password, "role", req.Role)
	user, err := c.service.GetUser(req.Email, req.Phone, req.Password)
	if err != nil {
		slog.Warn("Can't find user", "err", err)
		ctx.JSON(http.StatusNotFound, gin.H{"msg": "user was not found"})
		return
	}
	token, err := service.GenerateJWT(user)
	if err != nil {
		slog.Warn("Error generating jwt", "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "err"})
		return
	}
	ctx.JSON(http.StatusOK, dto.LoginResponseDTO{Token: token})
}
