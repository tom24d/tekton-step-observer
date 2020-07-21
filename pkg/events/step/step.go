package step

import "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

// TODO maybe it has to watch Pod and GET step container spec&status directly.
func GetSteps(taskrun *v1beta1.TaskRun) []v1beta1.Step {
	return taskrun.Status.TaskSpec.Steps
}

func GetStepStatuses(taskrun *v1beta1.TaskRun) []v1beta1.StepState {
	return taskrun.Status.Steps
}
