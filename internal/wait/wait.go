package wait

import (
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	CrudDelayTime = 1 * time.Second
	PollDelayTime = 2 * time.Second
	Timeout       = 20 * time.Minute
	MinTimeOut    = 3 * time.Second
)

var (
	PendingStatusCRUD          = []string{http.StatusText(http.StatusTooManyRequests), http.StatusText(http.StatusFailedDependency)}
	PendingStatusVerifyCreated = []string{http.StatusText(http.StatusFailedDependency), http.StatusText(http.StatusTooManyRequests), http.StatusText(http.StatusNotFound)}
	PendingStatusVerifyDeleted = []string{http.StatusText(200), http.StatusText(http.StatusTooManyRequests)}

	TargetStatusResourceChange  = []string{http.StatusText(http.StatusAccepted)}
	TargetStatusResourceExists  = []string{http.StatusText(200)}
	TargetStatusResourceDeleted = []string{http.StatusText(404)}
)

// GenerateWriteStateChangeConf creates configuration struct for the WaitForStateContext on resources undergoing write operation
func GenerateWriteStateChangeConf(fn resource.StateRefreshFunc) *resource.StateChangeConf {
	waitResourceWrite := &resource.StateChangeConf{
		Pending:    PendingStatusCRUD,
		Target:     TargetStatusResourceChange,
		Refresh:    fn,
		Timeout:    Timeout,
		Delay:      CrudDelayTime,
		MinTimeout: MinTimeOut,
	}
	return waitResourceWrite
}

// GenerateReadStateChangeConf creates configuration struct for the WaitForStateContext on resources undergoing read operation
func GenerateReadStateChangeConf(pending []string, target []string, fn resource.StateRefreshFunc) *resource.StateChangeConf {
	waitResourceRead := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    fn,
		Timeout:    Timeout,
		Delay:      PollDelayTime,
		MinTimeout: MinTimeOut,
	}
	return waitResourceRead
}
