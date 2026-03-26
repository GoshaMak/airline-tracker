package controller

import (
	"airline-tracker/internal/airport/command"
	"airline-tracker/internal/airport/dto"
	"airline-tracker/internal/airport/service"
	"airline-tracker/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type GateController struct {
	service *service.GateService
}

func NewGateController(i do.Injector) (*GateController, error) {
	return &GateController{
		service: do.MustInvoke[*service.GateService](i),
	}, nil
}

func RegisterGateRoutes(i do.Injector, r *gin.Engine) {
	c := do.MustInvoke[*GateController](i)
	g := r.Group("/admin", middleware.AuthMiddleware("admin"))
	{
		g.POST("/add_gate", c.AddGate)
		g.GET("/list_gates", c.ListGates)
	}
}

// @Summary new gate
// @Description creates a gate
// @Tags airport
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param gate body dto.CreateGateRequest true "gate info"
// @Success 201 "gate created"
// @Failure 400
// @Failure 500
// @Router /admin/add_gate [post]
func (c *GateController) AddGate(ctx *gin.Context) {
	req := &dto.CreateGateRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse args"})
		return
	}
	cmd, err := command.NewAddGateCommand(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}
	if err := c.service.AddGate(cmd); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "internal error"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "gate created"})
}

// @Summary list gates
// @Description returns all gates
// @Tags airport
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 "gate created"
// @Failure 400
// @Failure 500
// @Router /admin/list_gates [get]
func (c *GateController) ListGates(ctx *gin.Context) {}
