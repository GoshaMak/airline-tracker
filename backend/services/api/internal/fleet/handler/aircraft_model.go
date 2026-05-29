package handler

import (
	"api/internal/fleet/command"
	"api/internal/fleet/dto"
	"api/internal/fleet/usecase"
	"api/internal/middleware"
	userDomain "api/internal/user/domain"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	admin := r.Group("/admin", middleware.AuthMiddleware(userDomain.AdminRole))
	{
		admin.POST("/add_aircraft_model", h.AddAircraftModel)
		admin.GET("/aircraft_model/:id", h.AircraftModelInfo)
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
	const op = "AircraftModelHandler.AddAircraftModel"
	req := &dto.CreateAircraftModelRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	cmd, err := command.NewCreateAircraftModelCommand(req)
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	if err := h.uc.AddAircraftModel(cmd); err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"msg": "created"})
}

// @Summary get info (only admin)
// @Description aircraft model info
// @Tags Aircraft
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} dto.AircraftModelInfoResponse
// @Failure 400
// @Failure 401
// @Failure 404
// @Failure 500
// @Router /aircraft_model/{id} [get]
func (h *AircraftModelHandler) AircraftModelInfo(ctx *gin.Context) {
	const op = "AircraftModelHandler.AircraftModelInfo"
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusNotFound, gin.H{"msg": "not found"})
		return
	}

	am, err := h.uc.GetById(id)
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}

	resp, err := dto.ToResponseAircraftModelInfo(am)
	if err != nil {
		slog.Warn(op, "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
