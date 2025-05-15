package test

import (
	"pjt/root/internal/config"
	repository "pjt/root/internal/repository/dao"
	"pjt/root/internal/service"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService(t *testing.T) {
	config := &config.Configuration{}
	service := service.NewMyServcie(repository.NewMyDB(config), config)
	str := service.Test()
	assert.Equal(t, "hello world", str)
}
