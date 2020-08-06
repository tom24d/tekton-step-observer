package test

import (
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pipelineTest "github.com/tektoncd/pipeline/test"
)

// WaitForTaskRunState polls the status of the TaskRun called name from client every
// interval until inState returns `true` indicating it is done, returns an
// error or timeout. desc will be used to name the metric that is emitted to
// track how long it took for name to get into the state checked by inState.
func WaitForTaskRunState(c *clients, name string, inState pipelineTest.ConditionAccessorFn) error {
	return wait.PollImmediate(1*time.Second, 3*time.Minute, func() (bool, error) {
		r, err := c.TaskRunClient.Get(name, metav1.GetOptions{})
		if err != nil {
			return true, err
		}
		return inState(&r.Status)
	})
}
