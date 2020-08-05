package step

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

func Test_GetTime_Running(t *testing.T) {
	const step1 = "step1"
	want := time.Now()
	stepStateRunning := v1beta1.StepState{
		ContainerState: corev1.ContainerState{
			Running: &corev1.ContainerStateRunning{
				StartedAt: metav1.Time{
					Time: want,
				},
			},
		},
		Name: step1,
	}

	got, err := GetTime(&stepStateRunning, CloudEventTypeStepStarted)
	if err != nil {
		t.Fatalf("error occured: %v", err)
	}
	if !want.Equal(*got) {
		t.Fatalf("want time and got time have diff.")
	}
}


func Test_GetTime_Terminated(t *testing.T) {
	const step1 = "step1"
	now := time.Now()
	want := now.Add(time.Minute)
	stepStateRunning := v1beta1.StepState{
		ContainerState: corev1.ContainerState{
			Terminated: &corev1.ContainerStateTerminated{
				StartedAt: metav1.Time{
					Time: now,
				},
				FinishedAt: metav1.Time{
					Time: want,
				},
			},
		},
		Name: step1,
	}
	// succeeded
	got, err := GetTime(&stepStateRunning, CloudEventTypeStepSucceeded)
	if err != nil {
		t.Fatalf("error occured: %v", err)
	}
	if !want.Equal(*got) {
		t.Fatalf("want time and got time have diff.")
	}

	// failed
	got, err = GetTime(&stepStateRunning, CloudEventTypeStepFailed)
	if err != nil {
		t.Fatalf("error occured: %v", err)
	}
	if !want.Equal(*got) {
		t.Fatalf("want time and got time have diff.")
	}
}
