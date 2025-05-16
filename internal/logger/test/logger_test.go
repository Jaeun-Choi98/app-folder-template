package test

import (
	"pjt/internal/logger"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	logger, err := logger.NewCustomLogger("")
	assert.NoError(t, err)
	logger.Printf("test%s", " test")
}
