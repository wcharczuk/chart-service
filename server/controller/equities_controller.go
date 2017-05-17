package controller

import (
	"github.com/blendlabs/go-util"
	"github.com/blendlabs/go-web"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/chart-service/server/core"
	"github.com/wcharczuk/chart-service/server/model"
)

// Equities is the equities controller.
type Equities struct{}

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

func (e Equities) getAllHandler(rc *web.Ctx) web.Result {
	var all []model.Equity
	err := spiffy.Default().GetAll(&all)
	if err != nil {
		return rc.API().InternalError(err)
	}
	return rc.API().Result(all)
}

func (e Equities) searchHandler(rc *web.Ctx) web.Result {
	searchString, err := rc.RouteParam("query")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	results, err := model.SearchEquities(searchString)
	if err != nil {
		return rc.API().InternalError(err)
	}
	return rc.API().Result(results)
}

func (e Equities) getHandler(rc *web.Ctx) web.Result {
	id, err := rc.RouteParamInt("id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	var equity model.Equity
	err = spiffy.Default().GetByID(&equity, id)
	if err != nil {
		return rc.API().InternalError(err)
	}
	if equity.IsZero() {
		return rc.API().NotFound()
	}
	return rc.API().Result(equity)
}

func (e Equities) createHandler(rc *web.Ctx) web.Result {
	var equity model.Equity
	err := rc.PostBodyAsJSON(&equity)
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	err = spiffy.Default().Create(&equity)
	if err != nil {
		return rc.API().InternalError(err)
	}
	return rc.API().Result(equity)
}

func (e Equities) updateHandler(rc *web.Ctx) web.Result {
	id, err := rc.RouteParamInt("id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	var reference model.Equity
	err = spiffy.Default().GetByID(&reference, id)
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

	err = spiffy.Default().Update(&equity)
	if err != nil {
		return rc.API().InternalError(err)
	}
	return rc.API().Result(equity)
}

func (e Equities) patchHandler(rc *web.Ctx) web.Result {
	id, err := rc.RouteParamInt("id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	var reference model.Equity
	err = spiffy.Default().GetByID(&reference, id)
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

	err = spiffy.Default().Update(&reference)
	if err != nil {
		return rc.API().InternalError(err)
	}

	return rc.API().Result(reference)
}

func (e Equities) deleteHandler(rc *web.Ctx) web.Result {
	id, err := rc.RouteParamInt("id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	var equity model.Equity
	err = spiffy.Default().GetByID(&equity, id)
	if err != nil {
		return rc.API().InternalError(err)
	}
	if equity.IsZero() {
		return rc.API().NotFound()
	}

	err = spiffy.Default().Delete(equity)
	if err != nil {
		return rc.API().InternalError(err)
	}
	return rc.API().OK()
}
