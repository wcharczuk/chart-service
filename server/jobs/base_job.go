package jobs

import (
	"fmt"
	"time"

	"github.com/blendlabs/go-util"
)

type baseJob struct{}

func (job baseJob) Name() string {
	return "base_job"
}

func (job baseJob) OnStart() {
	job.logf("Job `%s` starting.", job.Name())
}

// OnComplete runs after the job body.
func (job baseJob) OnComplete(err error) {
	if err == nil {
		job.logf("Job `%s` complete.", job.Name())
	} else {
		job.logf("Job `%s` failed.", job.Name())
		job.error(err)
	}
}

func (job baseJob) logf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Printf("%s - %s.\n", util.Color(time.Now().UTC().Format(time.RFC3339), util.ColorGray), message)
}

func (job baseJob) error(err error) {
	message := fmt.Sprintf("%s:\n%v", util.Color("Exception", util.ColorRed), err)
	fmt.Printf("%s - %s.\n", util.Color(time.Now().UTC().Format(time.RFC3339), util.ColorGray), message)
}
