package core

import (
	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/go-web"
)

// ReadRouteValue reads a route value with a default.
func ReadRouteValue(rc *web.RequestContext, key, defaultValue string) string {
	if value, err := rc.RouteParameter(key); err == nil {
		return value
	}
	return defaultValue
}

// ReadRouteValueInt reads a route value with a default.
func ReadRouteValueInt(rc *web.RequestContext, key string, defaultValue int) int {
	if value, err := rc.RouteParameterInt(key); err == nil {
		return value
	}
	return defaultValue
}

// ReadQueryValue reads a query value with a default.
func ReadQueryValue(rc *web.RequestContext, key, defaultValue string) string {
	if value, err := rc.QueryParam(key); err == nil {
		return value
	}
	return defaultValue
}

// ReadQueryValueInt reads a query value with a default.
func ReadQueryValueInt(rc *web.RequestContext, key string, defaultValue int) int {
	if value, err := rc.QueryParamInt(key); err == nil {
		return value
	}
	return defaultValue
}

// ReadQueryValueFloat64 reads a query value with a default.
func ReadQueryValueFloat64(rc *web.RequestContext, key string, defaultValue float64) float64 {
	if value, err := rc.QueryParamFloat64(key); err == nil {
		return value
	}
	return defaultValue
}

// ReadQueryValueBool reads a query value with a default.
func ReadQueryValueBool(rc *web.RequestContext, key string, defaultValue bool) bool {
	if value, err := rc.QueryParam(key); err == nil {
		return util.CaseInsensitiveEquals(value, "true")
	}
	return defaultValue
}
