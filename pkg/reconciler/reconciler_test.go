package reconciler

import (
	"github.com/tom24d/step-observe-controller/pkg/events/step"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/test/diff"
)

func Test_initializeAnnotation(t *testing.T) {
	stepName1 := "step-1"
	stepName2 := "step-2"
	trstatus := v1beta1.TaskRunStatus{
		TaskRunStatusFields: v1beta1.TaskRunStatusFields{
			PodName:        "pod-name",
			StartTime:      nil,
			CompletionTime: nil,
			Steps: []v1beta1.StepState{
				v1beta1.StepState{
					ContainerState: corev1.ContainerState{},
					Name:           stepName1,
					ContainerName:  "step-container-1",
					ImageID:        "",
				},
				v1beta1.StepState{
					ContainerState: corev1.ContainerState{},
					Name:           stepName2,
					ContainerName:  "step-container-2",
					ImageID:        "",
				},
			},
			TaskSpec: &v1beta1.TaskSpec{
				Steps: []v1beta1.Step{
					v1beta1.Step{
						Container: corev1.Container{
							Name: stepName1,
						},
						Script: "",
					},
					v1beta1.Step{
						Container: corev1.Container{
							Name: stepName2,
						},
						Script: "",
					},
				},
			},
		},
	}
	tr := v1beta1.TaskRun{
		Status: trstatus,
	}
	wantEmissionStatuses := &step.EmissionStatuses{
		Statuses: []step.EmissionStatus{
			{
				Name:    stepName1,
				Emitted: []step.TektonPluginEventType{},
			},
			{
				Name:    stepName2,
				Emitted: []step.TektonPluginEventType{},
			},
		},
	}

	got, err := initializeAnnotation(&tr)
	if err != nil {
		t.Fatal(err)
	}
	if d := cmp.Diff(wantEmissionStatuses, got); d != "" {
		t.Fatalf("Annotation data mismatch: %s", diff.PrintWantGot(d))
	}
}
