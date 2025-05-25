package test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"pjt/internal/container"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestREST(t *testing.T) {
	container, err := container.NewContainer()
	assert.NoError(t, err)

	ms := httptest.NewServer(container.Controller.Router)
	defer ms.Close()

	req, err := http.NewRequest("", ms.URL+"/test", nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	resBytes, err := io.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, "hello world", string(resBytes))
}
