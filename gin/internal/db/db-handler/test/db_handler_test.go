package test

import (
	"pjt/internal/container"
	dbhandler "pjt/internal/db/db-handler"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingOracle(t *testing.T) {
	container, err := container.NewContainer()
	assert.NoError(t, err)
	t.Log(container.Config.DBUser, container.Config.DBPasswd)
	dbHandler, err := dbhandler.NewDBHandler(container.Config)
	assert.NoError(t, err)
	defer dbHandler.Close()
	assert.NoError(t, dbHandler.Ping())
}
