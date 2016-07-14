package controller

import (
	"github.com/blendlabs/go-chronometer"
	"github.com/wcharczuk/go-web"
)

type Jobs struct{}

// GET "/api/v1/jobs"
func (j Jobs) getJobsAction(rc *web.RequestContext) web.ControllerResult {
	status := chronometer.Default().Status()
	return rc.API().JSON(status)
}

// POST "/api/v1/job/:job_id"
func (j Jobs) runJobAction(rc *web.RequestContext) web.ControllerResult {
	jobID, err := rc.RouteParameter("job_id")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}

	err = chronometer.Default().RunJob(jobID)
	if err != nil {
		return rc.API().InternalError(err)
	}
	return rc.API().OK()
}

// Register registers the controllers routes.
func (j Jobs) Register(app *web.App) {
	app.GET("/api/v1/jobs", j.getJobsAction)
	app.POST("/api/v1/job/:job_id", j.runJobAction)
}
