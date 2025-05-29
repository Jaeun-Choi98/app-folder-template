package middleware

import (
	"net/http"
	"pjt/internal/logger"
	"pjt/internal/transport/http-rest/http-utils/httperr"

	"github.com/gin-gonic/gin"
)

func ErrorMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		for _, ctxErr := range ctx.Errors {
			if err, ok := ctxErr.Err.(*httperr.HttpError); ok {
				if err.Code == http.StatusInternalServerError {
					logger.Println(err.ErrMsg)
				}
				ctx.JSON(err.Code, err)
			}
		}
	}
}
