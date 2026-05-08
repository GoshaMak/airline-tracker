package handler

import (
	"airline-tracker/internal/common"
	"airline-tracker/internal/fleet/command"
	"airline-tracker/internal/fleet/dto"
	"airline-tracker/internal/fleet/usecase"
	"airline-tracker/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type AircraftModelHandler struct {
	uc *usecase.AircraftModelUsecase
}

func NewAircraftModelHandler(i do.Injector) (*AircraftModelHandler, error) {
	return &AircraftModelHandler{
		uc: do.MustInvoke[*usecase.AircraftModelUsecase](i),
	}, nil
}

func RegisterAircraftModelRoutes(i do.Injector, r *gin.Engine) {
	h := do.MustInvoke[*AircraftModelHandler](i)

	admin := r.Group("/admin", middleware.AuthMiddleware(common.AdminRole))
	{
		admin.POST("/add_aircraft_model", h.AddAircraftModel)
	}
}

// @Summary add aircraft model
// @Description create new aircraft model
// @Tags Aircraft
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param aircraft_model body dto.CreateAircraftModelRequest true "aircraft model info"
// @Success 201 "created"
// @Failure 400
// @Failure 401
// @Router /admin/add_aircraft_model [post]
func (h *AircraftModelHandler) AddAircraftModel(ctx *gin.Context) {
	req := &dto.CreateAircraftModelRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse args", "err": err})
		return
	}
	cmd, err := command.NewCreateAircraftModelCommand(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid args", "err": err})
		return
	}
	if err := h.uc.AddAircraftModel(cmd); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request", "err": err})
		return
	}
	ctx.JSON(http.StatusCreated, "created")
}
