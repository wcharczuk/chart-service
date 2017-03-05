package controller

import (
	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-web"
)

// Jobs is a controller that exposes methods to control jobs.
type Jobs struct{}

// GET "/api/v1/jobs"
func (j Jobs) getJobsAction(rc *web.Ctx) web.Result {
	status := chronometer.Default().Status()
	return rc.API().Result(status)
}

// POST "/api/v1/job/:job_id"
func (j Jobs) runJobAction(rc *web.Ctx) web.Result {
	jobID, err := rc.RouteParam("job_id")
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
