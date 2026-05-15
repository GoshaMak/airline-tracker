package handler

import (
	"api/internal/fleet/command"
	"api/internal/fleet/dto"
	"api/internal/fleet/usecase"
	"api/internal/middleware"
	userDomain "api/internal/user/domain"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type AircraftHandler struct {
	uc *usecase.AircraftUsecase
}

func NewAircraftHandler(i do.Injector) (*AircraftHandler, error) {
	return &AircraftHandler{
		uc: do.MustInvoke[*usecase.AircraftUsecase](i),
	}, nil
}

func RegisterAircraftRoutes(i do.Injector, r *gin.Engine) {
	h := do.MustInvoke[*AircraftHandler](i)

	admin := r.Group("/admin", middleware.AuthMiddleware(userDomain.AdminRole))
	{
		admin.POST("/add_aircraft", h.AddAircraft)
	}
}

// @Summary add aircraft
// @Description create new aircraft
// @Tags Aircraft
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param aircraft body dto.CreateAircraftRequest true "aircraft info"
// @Success 201 "created"
// @Failure 400
// @Failure 401
// @Router /admin/add_aircraft [post]
func (h *AircraftHandler) AddAircraft(ctx *gin.Context) {
	req := &dto.CreateAircraftRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse args"})
		return
	}
	cmd, err := command.NewCreateAircraftCommand(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	if err := h.uc.AddAircraft(cmd); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	ctx.JSON(http.StatusCreated, "created")
}
