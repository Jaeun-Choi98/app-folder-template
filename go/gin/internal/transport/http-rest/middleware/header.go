package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func ApkHeader(fname string, fsize int) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/vnd.android.package-archive")
		// no-cache: 재검증, no-store: 캐시금지
		c.Header("Cache-Control", "no-cache")
		// 보안
		c.Header("X-Content-Type-Options", "nosniff")
		// attachment: 브라우저에게 "화면에 표시하지 말고 파일로 저장하라"고 지시합니다.
		// 이게 없으면 브라우저가 내용을 열어 보려고 할 수 있습니다.
		// filename: 저장할 파일 이름을 알려 줍니다.
		// 없으면 클라이언트가 URL 마지막 경로(예: /download/apk라면 apk)로 이름을 지어서,
		// 확장자 없는 파일이나 엉뚱한 이름으로 저장될 수 있습니다. 버전별 파일명을 서버가 정해 주고 싶을 때도 필요합니다.
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fname))
		// 진행률 표시: 클라이언트가 전체 크기를 알아야 “45% 다운로드됨” 같은 진행 상황과 남은 시간을 계산할 수 있습니다. 없으면 진행률을 알 수 없는 다운로드가 됩니다.
		// 완료 여부 확인: 받은 바이트 수와 비교해서 다운로드가 중간에 끊겼는지 판단할 수 있습니다. APK는 몇 바이트만 빠져도 설치가 실패하니 특히 중요합니다.
		// 전송 방식 결정: 크기를 모르면 HTTP/1.1이 chunked 방식으로 전송합니다. 크기를 주면 일반 전송이 되고, 이어받기(Range)에도 유리합니다.
		c.Header("Content-Length", fmt.Sprintf("%d", fsize))
		// custom header
		c.Header("file-size", fmt.Sprintf("%d", fsize))
		c.Next()
	}
}

// 4. 전송
// c.FileAttachment(path, name)
