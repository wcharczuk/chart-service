package core

import (
	"net/http"
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-web"
)

func TestAuthRequiredProd(t *testing.T) {
	t.Skip()
	assert := assert.New(t)

	Config.env = "prod"
	Config.authKey = "test_key"

	app := web.New()
	app.GET("/", func(rc *web.Ctx) web.Result {
		return rc.API().OK()
	}, AuthRequired, web.APIProviderAsDefault)

	meta, err := app.Mock().WithPathf("/").WithHeader(AuthKeyParamName, "test_key").ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusOK, meta.StatusCode)

	meta, err = app.Mock().WithPathf("/").WithHeader(AuthKeyParamName, "not_test_key").ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusForbidden, meta.StatusCode)

	meta, err = app.Mock().WithPathf("/").ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusForbidden, meta.StatusCode)
}

func TestAuthRequiredDev(t *testing.T) {
	t.Skip()
	assert := assert.New(t)

	Config.env = "dev"
	Config.authKey = "test_key"

	app := web.New()
	app.GET("/", func(rc *web.Ctx) web.Result {
		return rc.API().OK()
	}, AuthRequired, web.APIProviderAsDefault)

	meta, err := app.Mock().WithPathf("/").WithHeader(AuthKeyParamName, "test_key").ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusOK, meta.StatusCode)

	meta, err = app.Mock().WithPathf("/").WithHeader(AuthKeyParamName, "not_test_key").ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusOK, meta.StatusCode)
}
