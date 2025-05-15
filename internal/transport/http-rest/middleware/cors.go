package middleware

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func NewCORSMiddleware(origins []string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", strings.Join(origins, ","))
			w.Header().Add("Access-Control-Allow-Headers",
				"Content-Type,AccessToken,X-CSRF-Token,Authorization,Token,Set-Cookie,X-Requested-With")
			w.Header().Add("Access-Control-Allow-Credentials", "true")
			w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			//w.Header().Set("content-type", "application/json;charset=UTF-8")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
