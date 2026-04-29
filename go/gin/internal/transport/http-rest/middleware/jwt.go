package middleware

import (
	"fmt"
	"pjt/internal/infra/cache"
	"pjt/internal/transport/http-rest/http-utils/httperr"
	"pjt/internal/transport/http-rest/http-utils/jwt"
	"pjt/internal/transport/http-rest/response"
	"slices"
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

func CheckSession() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !slices.Contains([]string{"POST", "PUT", "PATCH"}, ctx.Request.Method) {
			ctx.Next()
			return
		}
		accToken := ctx.GetHeader("Authorization")
		if accToken != "" {
			accToken, _ = strings.CutPrefix(accToken, "Bearer ")
			if claims, err := jwt.VaildJwtHS256(accToken); err == nil {
				if cache.CheckAccToken(claims.SampleModelId, accToken) {
					ctx.Next()
					return
				}
			}
		}
		// var token string
		// if cookie, err := ctx.Request.Cookie("refToken"); err == nil {
		// 	log.Println(cookie.Value)
		// 	if token = cookie.Value; token != "" {
		// 		token, _ = strings.CutPrefix(token, "Bearer ")
		// 		if claims, err := jwt.VaildJwtHS256(token); err == nil {
		// 			if service.CheckRefToken(claims.Id, token) {
		// 				ctx.Next()
		// 				return
		// 			}
		// 		}
		// 	}
		// }
		ctx.Error(httperr.UNAUTHORIZED.Add(fmt.Errorf("[REST] CheckSession, invalid token"), response.INVALID_TOKEN))
		ctx.Abort()
	}
}
