package controller

import (
	"github.com/blendlabs/go-util"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/chart-service/server/core"
	"github.com/wcharczuk/chart-service/server/model"
	"github.com/wcharczuk/go-web"
)

// Equities is the equities controller.
type Equities struct{}

func (e Equities) getAllHandler(rc *web.RequestContext) web.ControllerResult {
	var all []model.Equity
	err := spiffy.DefaultDb().GetAll(&all)
	if err != nil {
		return rc.API().InternalError(err)
	}
	return rc.API().JSON(all)
}

func (e Equities) searchHandler(rc *web.RequestContext) web.ControllerResult {
	searchString, err := rc.RouteParameter("query")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	results, err := model.SearchEquities(searchString)
	if err != nil {
		return rc.API().InternalError(err)
	}
	return rc.API().JSON(results)
}

func (e Equities) getHandler(rc *web.RequestContext) web.ControllerResult {
	id, err := rc.RouteParameterInt("id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	var equity model.Equity
	err = spiffy.DefaultDb().GetByID(&equity, id)
	if err != nil {
		return rc.API().InternalError(err)
	}
	if equity.IsZero() {
		return rc.API().NotFound()
	}
	return rc.API().JSON(equity)
}

func (e Equities) createHandler(rc *web.RequestContext) web.ControllerResult {
	var equity model.Equity
	err := rc.PostBodyAsJSON(&equity)
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	err = spiffy.DefaultDb().Create(&equity)
	if err != nil {
		return rc.API().InternalError(err)
	}
	return rc.API().JSON(equity)
}

func (e Equities) updateHandler(rc *web.RequestContext) web.ControllerResult {
	id, err := rc.RouteParameterInt("id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	var reference model.Equity
	err = spiffy.DefaultDb().GetByID(&reference, id)
	if err != nil {
		return rc.API().InternalError(err)
	}
	if reference.IsZero() {
		return rc.API().NotFound()
	}

	var equity model.Equity
	err = rc.PostBodyAsJSON(&equity)
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}

	equity.ID = reference.ID

	err = spiffy.DefaultDb().Update(&equity)
	if err != nil {
		return rc.API().InternalError(err)
	}
	return rc.API().JSON(equity)
}

func (e Equities) patchHandler(rc *web.RequestContext) web.ControllerResult {
	id, err := rc.RouteParameterInt("id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	var reference model.Equity
	err = spiffy.DefaultDb().GetByID(&reference, id)
	if err != nil {
		return rc.API().InternalError(err)
	}
	if reference.IsZero() {
		return rc.API().NotFound()
	}

	var patchData map[string]interface{}
	err = rc.PostBodyAsJSON(&patchData)
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}

	err = util.Reflection.PatchObject(&reference, patchData)
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}

	err = spiffy.DefaultDb().Update(&reference)
	if err != nil {
		return rc.API().InternalError(err)
	}

	return rc.API().JSON(reference)
}

func (e Equities) deleteHandler(rc *web.RequestContext) web.ControllerResult {
	id, err := rc.RouteParameterInt("id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	var equity model.Equity
	err = spiffy.DefaultDb().GetByID(&equity, id)
	if err != nil {
		return rc.API().InternalError(err)
	}
	if equity.IsZero() {
		return rc.API().NotFound()
	}

	err = spiffy.DefaultDb().Delete(equity)
	if err != nil {
		return rc.API().InternalError(err)
	}
	return rc.API().OK()
}

// Register registers the controller.
func (e Equities) Register(app *web.App) {
	app.GET("/api/v1/equities", e.getAllHandler)
	app.GET("/api/v1/equities.search/:query", e.searchHandler)
	app.POST("/api/v1/equity", e.createHandler, core.AuthRequired, web.APIProviderAsDefault)
	app.GET("/api/v1/equity/:id", e.getHandler)
	app.PUT("/api/v1/equity/:id", e.updateHandler, core.AuthRequired, web.APIProviderAsDefault)
	app.PATCH("/api/v1/equity/:id", e.patchHandler, core.AuthRequired, web.APIProviderAsDefault)
	app.DELETE("/api/v1/equity/:id", e.deleteHandler, core.AuthRequired, web.APIProviderAsDefault)
}
