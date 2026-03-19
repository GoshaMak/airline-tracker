package controller

import (
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
		g.POST("/gate", c.AddGate)
	}
}

func (c *GateController) AddGate(ctx *gin.Context) {
	type Req struct {
		Gate dto.GateDTO `json:"gate"`
	}
	req := &Req{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse args"})
		return
	}
	if err := c.service.AddGate(req.Gate.GateFromDTO()); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "gate created"})
}
