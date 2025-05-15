package test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"pjt/internal/config"
	repository "pjt/internal/repository/dao"
	"pjt/internal/service"
	"pjt/internal/transport/http-rest/controller"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestREST(t *testing.T) {
	config := &config.Configuration{}
	controller := controller.NewController(service.NewMyServcie(repository.NewMyDB(config), config), config)
	ms := httptest.NewServer(controller.Router)
	defer ms.Close()
	req, err := http.NewRequest("", ms.URL+"/test", nil)
	assert.NoError(t, err)
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	resBytes, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "hello world", string(resBytes))
}
