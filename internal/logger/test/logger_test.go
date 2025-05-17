package test

import (
	"pjt/internal/logger"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	customLogger, err := logger.NewCustomLogger("")
	logger.SetLogger(customLogger)

	logger.StartCleaning()
	defer logger.StopCleaning()
	time.Sleep(time.Second * 3)

	assert.NoError(t, err)
	logger.Printf("test%s", " test")

}
