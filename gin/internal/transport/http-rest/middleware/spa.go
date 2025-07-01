package middleware

import (
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

/**
 * Gin을 사용한 spa 구현, gin은 미들웨어를 통해 url을 검증하고 서비스를 제공.
 */
func SpaHandlerRoot(staticPath, indexPath string) gin.HandlerFunc {

	return func(c *gin.Context) {
		log.Println("root")
		url := c.Request.URL.Path
		ext := filepath.Ext(url)

		// 1. 민감한 확장자 차단
		blockedExt := map[string]bool{
			".git": true,
			".ini": true,
			".svg": true,
			".txt": true,
		}
		if blockedExt[ext] {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// 2. 경로 정리 및 탐색 방지
		cleanPath := filepath.Clean(url)
		absPath := filepath.Join(staticPath, cleanPath)
		_, err := os.Stat(absPath)
		/**
		 * err != nil 일 때,
		 * 1. 파일이 존재x  -> fallback
		 * 2. 파일에 접근 권한이 없음
		 * 3. 드라이브가 연결x
		 * 4. 깨진 심볼릭 링크를 참조
		 *
		 * 1번의 경우에만 Fallback
		 */
		if err != nil {
			// 1. 파일이 존재하지x
			if errors.Is(err, fs.ErrNotExist) {
				// fallback
				http.ServeFile(c.Writer, c.Request, filepath.Join(staticPath, indexPath))
				return
			} else {
				http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		http.FileServer(http.Dir(staticPath)).ServeHTTP(c.Writer, c.Request)
	}
}

func SpaHandlerOther(urlPrefix, staticPath string) gin.HandlerFunc {
	return func(c *gin.Context) {

		log.Println("other")
		url := c.Request.URL.Path
		ext := filepath.Ext(url)

		// 1. 민감한 확장자 차단
		blockedExt := map[string]bool{
			".git": true,
			".ini": true,
			".svg": true,
			".txt": true,
		}
		if blockedExt[ext] {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// 2. 경로 정리 및 탐색 방지
		cleanPath := filepath.Clean(url)
		// if req url: "/example/img/aa.jpg", urlPrefix:"/example" -> search in fileserver /img/aa.jpg
		cleanPath, hasPrefix := strings.CutPrefix(cleanPath, filepath.FromSlash(urlPrefix))

		// if url: "/img/aa.jpg", urlPrefix:"/example" -> not processing this middleware, passing next middleware
		if !hasPrefix {
			c.Next()
			return
		}

		absPath := filepath.Join(staticPath, cleanPath)

		_, err := os.Stat(absPath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			} else {
				http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
				c.Abort()
				return
			}
		}

		/**
		 * StripPrefix는 c.Request.URL.Path에서 urlPrefix를 제거하고 ServeHTTP를 제공함.
		 * 추가적으로 StripPrefix는 urlPrefix로 시작하지 않으면, 404에러를 띄움.
		 */
		http.StripPrefix(urlPrefix, http.FileServer(http.Dir(staticPath))).ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}
