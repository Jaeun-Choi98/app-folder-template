package middleware

import (
	"log"
	"net/http"
	"time"
)

func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		before := time.Now()
		next.ServeHTTP(w, r)
		elapsed := time.Now().Sub(before)
		log.Printf("%s -> [%s]%s: %v sec", r.RemoteAddr, r.Method, r.URL.Path, elapsed.Seconds())
	})
}
