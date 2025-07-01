package middleware

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

/**
 * Gin을 사용한 spa 구현, gin은 미들웨어를 통해 url을 검증하고 서비스를 제공.
 */
func SpaHandlerGin(staticPath, indexPath string) gin.HandlerFunc {

	return func(c *gin.Context) {
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
