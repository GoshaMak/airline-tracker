package route

import (
	controllers "airline-ticketing-svc/internal/controller"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

func RegisterUserRoutes(i do.Injector, r *gin.Engine) {
	uc := do.MustInvoke[*controllers.UserController](i)
	r.GET("/login", uc.Login)
	r.POST("/signup", uc.SignUp)
}
