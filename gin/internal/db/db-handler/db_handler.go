package dbhandler

import (
	"errors"
	"fmt"
	"pjt/internal/config"
	"pjt/internal/db/db-handler/oracle"
	"strings"
	"time"
)

type DBHandlerInterface interface {
	Test() string
	Close() error
	Ping() error
	StartDBHeartbeat()
	StopDBHeartbeat()
}

func NewDBHandler(config *config.Configuration) (DBHandlerInterface, error) {

	switch config.DBType {
	case "oracle":
		return oracle.NewOralce(getOracleDSN(config), time.Duration(config.DBHeartbeat)*time.Second)
	case "maria":
		return nil, nil
	case "postgres":
		return nil, nil
	case "tibero":
		return nil, nil
	default:
		return nil, errors.New("not exsits database type")
	}
}

func getPostgresDSN(c *config.Configuration) string {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		c.DBUser, c.DBPasswd, c.DBHost, c.DBPort, c.DBName)

	params := make([]string, 0)

	// SSL 모드 (기본값: disable)
	// sslMode := c.SSLMode
	// if sslMode == "" {
	// 	sslMode = "disable"
	// }
	// params = append(params, "sslmode="+sslMode)

	// 연결 타임아웃
	if c.DBConnectTimeout > 0 {
		params = append(params, fmt.Sprintf("connect_timeout=%d", c.DBConnectTimeout))
	}

	// 시간대 설정
	timeZone := c.DBTimezone
	if timeZone == "" {
		timeZone = "Local" // 기본값
	}
	params = append(params, "timezone="+timeZone)

	// 애플리케이션 이름
	// if c.ApplicationName != "" {
	// 	params = append(params, "application_name="+c.ApplicationName)
	// }

	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}

	return dsn
}

func getMysqlOrMariaDSN(c *config.Configuration) string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		c.DBUser, c.DBPasswd, c.DBHost, c.DBPort, c.DBName)

	params := make([]string, 0)

	// Charset 설정 (기본값: utf8mb4 권장)
	charset := c.DBCharset
	if charset == "" {
		charset = "utf8mb4" // 이모지 지원
	}
	params = append(params, "charset="+charset)

	// ParseTime 설정 (time.Time 사용시 필요)
	if c.ParseTime {
		params = append(params, "parseTime=true")
	}

	// 시간대 설정 (loc 파라미터)
	timeZone := c.DBTimezone
	if timeZone == "" {
		timeZone = "Local" // 기본값
	}
	params = append(params, "loc="+timeZone)

	// 연결 타임아웃
	if c.DBConnectTimeout > 0 {
		params = append(params, fmt.Sprintf("timeout=%ds", c.DBConnectTimeout))
	}

	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}

	return dsn
}

func getOracleDSN(c *config.Configuration) string {
	// user/password@host:port/service_name?param1=value1&param2=value2
	connectString := fmt.Sprintf("%s:%s/%s", c.DBHost, c.DBPort, c.ServiceName)
	dsn := fmt.Sprintf("%s/%s@%s", c.DBUser, c.DBPasswd, connectString)

	params := make([]string, 0)

	/**
	 * Charset 설정 (UTF-8이 기본값, 대부분 불필요)
	 * 참고: godror 기본값은 charset=UTF-8, 특별한 이유 없다면 설정 불필요
	 */
	// if c.DBCharset != "" {
	//     params = append(params, "charset="+c.Charset)
	// }

	// 시간대 설정
	timeZone := c.DBTimezone
	if timeZone == "" {
		timeZone = "Local" // 기본값
	}
	params = append(params, "timezone="+timeZone)

	// 연결 타임아웃
	if c.DBConnectTimeout > 0 {
		params = append(params, fmt.Sprintf("connect_timeout=%ds", c.DBConnectTimeout))
	}

	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}

	return dsn
}

func getODBCDSN(c *config.Configuration) string {
	if c.DSN != "" {
		// 미리 설정된 DSN 사용
		return fmt.Sprintf("DSN=%s;UID=%s;PWD=%s", c.DSN, c.DBUser, c.DBPasswd)
	}

	// 직접 연결 문자열 구성
	parts := []string{
		fmt.Sprintf("DRIVER={%s}", c.Driver),
		fmt.Sprintf("SERVER=%s", c.DBHost),
		fmt.Sprintf("DATABASE=%s", c.DBName),
		fmt.Sprintf("UID=%s", c.DBUser),
		fmt.Sprintf("PWD=%s", c.DBPasswd),
	}

	if c.DBPort != "" {
		parts = append(parts, fmt.Sprintf("PORT=%s", c.DBPort))
	}

	// 연결 타임아웃 (드라이버에 따라 키워드 다름)
	if c.DBConnectTimeout > 0 {
		parts = append(parts, fmt.Sprintf("CONNECTION_TIMEOUT=%d", c.DBConnectTimeout))
	}

	// Charset 설정 (드라이버에 따라 다름)
	if c.DBCharset != "" {
		parts = append(parts, "CHARSET="+c.DBCharset)
	}

	return strings.Join(parts, ";")
}
