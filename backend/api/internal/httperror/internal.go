package httperror

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Internal(ctx *gin.Context, err, denied error) {
	body := gin.H{"msg": "internal error"}
	if denied != nil && errors.Is(err, denied) {
		body = gin.H{
			"code": "operation_not_permitted",
			"msg":  "operation is not permitted",
		}
	}
	ctx.AbortWithStatusJSON(http.StatusInternalServerError, body)
}
