package middleware

import (
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
			ctx.AbortWithError(httperr.UNAUTHORIZED_CODE, httperr.UNAUTHORIZED)
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
			ctx.AbortWithError(httperr.UNAUTHORIZED_CODE, httperr.UNAUTHORIZED)
			return
		}

		if claims.Name == "" {
			ctx.AbortWithError(httperr.UNAUTHORIZED_CODE, httperr.UNAUTHORIZED)
			return
		}

		ctx.Next()
	}
}

func StoreIdToContext() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var token string
		if cookie, err := ctx.Request.Cookie("jwt"); err == nil {
			if token = cookie.Value; token != "" {
				token, _ = strings.CutPrefix(token, "Bearer ")
				if claims, err := jwt.VaildJwtHS256(token); err == nil {
					ctx.Set("id", claims.SampleModelId)
				}
			}
		}
		ctx.Next()
	}
}
