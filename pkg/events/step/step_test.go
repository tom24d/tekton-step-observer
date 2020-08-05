package step

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/google/go-cmp/cmp"

	"github.com/tektoncd/pipeline/test/diff"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

func Test_GetSteps(t *testing.T) {

	want := []v1beta1.Step{
		v1beta1.Step{
			Container: corev1.Container{
				Name: "step-container-1",
			},
		},
	}

	run := &v1beta1.TaskRun{
		Status: v1beta1.TaskRunStatus{
			TaskRunStatusFields: v1beta1.TaskRunStatusFields{
				TaskSpec: &v1beta1.TaskSpec{
					Steps: want,
				},
			},
		},
	}

	got := GetSteps(run)

	if d := cmp.Diff(want, got); d != "" {
		t.Fatalf(diff.PrintWantGot(d))
	}
}

func Test_GetStepStatuses(t *testing.T) {

	const step1 = "step1"
	const step2 = "step2"

	want := []v1beta1.StepState{
		v1beta1.StepState{
			Name: step1,
		},
		v1beta1.StepState{
			Name: step2,
		},
	}

	run := &v1beta1.TaskRun{
		Status: v1beta1.TaskRunStatus{
			TaskRunStatusFields: v1beta1.TaskRunStatusFields{
				Steps: want,
			},
		},
	}

	got := GetStepStatuses(run)

	if d := cmp.Diff(want, got); d != "" {
		t.Fatalf(diff.PrintWantGot(d))
	}
}
