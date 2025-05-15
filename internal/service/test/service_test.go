package test

import (
	"pjt/internal/config"
	repository "pjt/internal/repository/dao"
	"pjt/internal/service"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService(t *testing.T) {
	config := &config.Configuration{}
	service := service.NewMyServcie(repository.NewMyDB(config), config)
	str := service.Test()
	assert.Equal(t, "hello world", str)
}
