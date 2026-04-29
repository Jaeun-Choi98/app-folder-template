package test

import (
	"pjt/internal/container"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService(t *testing.T) {
	container, err := container.NewContainer()
	assert.NoError(t, err)

	str := container.Service.Test()
	assert.Equal(t, "hello world", str)
}
