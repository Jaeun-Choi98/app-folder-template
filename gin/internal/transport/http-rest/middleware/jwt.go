package middleware

import (
	"fmt"
	"pjt/internal/transport/http-rest/http-utils/httperr"
	"pjt/internal/transport/http-rest/http-utils/jwt"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// 쿠키를 사용하는 경우
		cookie, err := ctx.Request.Cookie("jwt")
		if err != nil {
			ctx.Error(httperr.UNAUTHORIZED.AddErrMsg(err))
			ctx.Abort()
			return
		}
		jwtStr := cookie.Value
		/*
			헤더를 사용하는 경우
			jwtStr := ctx.GetHeader("Authorization")
			if jwtStr == "" {
				ctx.Error(httperr.UNAUTHORIZED.AddErrMsg(err))
				ctx.Abort()
				return
			}
		*/
		jwtStr, _ = strings.CutPrefix(jwtStr, "Bearer ")
		claims, err := jwt.VaildJwtHS256(jwtStr)
		if err != nil {
			ctx.Error(httperr.UNAUTHORIZED.AddErrMsg(err))
			ctx.Abort()
			return
		}

		if claims.Name == "" {
			ctx.Error(httperr.UNAUTHORIZED.AddErrMsg(fmt.Errorf("claim name is empty")))
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
