package core

import (
	"crypto/hmac"

	"github.com/blendlabs/go-util"
	"github.com/blendlabs/go-web"
)

const (
	//AuthKeyParamName is the param name for the auth key
	AuthKeyParamName = "auth"
)

// AuthRequired is a special type of middleware that checks an `auth` header param.
func AuthRequired(action web.Action) web.Action {
	return func(context *web.Ctx) web.Result {
		if context.DefaultResultProvider() == nil {
			panic("You must provide a content provider as middleware to use `AuthRequired`")
		}

		if Config.IsProduction() {
			authKey := context.Param(AuthKeyParamName)

			if !hmac.Equal([]byte(authKey), []byte(Config.AuthKey())) {
				return context.DefaultResultProvider().NotAuthorized()
			}

		}
		return action(context)
	}
}

// ReadRouteValue reads a route value with a default.
func ReadRouteValue(rc *web.Ctx, key, defaultValue string) string {
	if value, err := rc.RouteParam(key); err == nil {
		return value
	}
	return defaultValue
}

// ReadRouteValueInt reads a route value with a default.
func ReadRouteValueInt(rc *web.Ctx, key string, defaultValue int) int {
	if value, err := rc.RouteParamInt(key); err == nil {
		return value
	}
	return defaultValue
}

// ReadQueryValue reads a query value with a default.
func ReadQueryValue(rc *web.Ctx, key, defaultValue string) string {
	if value, err := rc.QueryParam(key); err == nil {
		return value
	}
	return defaultValue
}

// ReadQueryValueInt reads a query value with a default.
func ReadQueryValueInt(rc *web.Ctx, key string, defaultValue int) int {
	if value, err := rc.QueryParamInt(key); err == nil {
		return value
	}
	return defaultValue
}

// ReadQueryValueFloat64 reads a query value with a default.
func ReadQueryValueFloat64(rc *web.Ctx, key string, defaultValue float64) float64 {
	if value, err := rc.QueryParamFloat64(key); err == nil {
		return value
	}
	return defaultValue
}

// ReadQueryValueBool reads a query value with a default.
func ReadQueryValueBool(rc *web.Ctx, key string, defaultValue bool) bool {
	if value, err := rc.QueryParam(key); err == nil {
		return util.String.CaseInsensitiveEquals(value, "true")
	}
	return defaultValue
}
