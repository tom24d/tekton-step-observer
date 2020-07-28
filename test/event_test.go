//+build e2e

package test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cetestv2 "github.com/cloudevents/sdk-go/v2/test"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	"github.com/tom24d/step-observe-controller/pkg/events/step/resources"
)

func Test_EventAssertion(t *testing.T) {

	successStep := v1beta1.Step{Container: corev1.Container{
		Image:   "busybox@sha256:895ab622e92e18d6b461d671081757af7dbaa3b00e3e28e12505af7817f73649",
		Command: []string{"/bin/sh"},
		Args:    []string{"-c", "echo hello1"},
	},
	}

	failStep := v1beta1.Step{Container: corev1.Container{
		Image:   "busybox@sha256:895ab622e92e18d6b461d671081757af7dbaa3b00e3e28e12505af7817f73649",
		Command: []string{"/bin/sh"},
		Args:    []string{"-c", "exit 1"},
	},
	}

	assertSetGetFunc := func(event resources.TektonPluginEventType) cetestv2.EventMatcher {
		return cetestv2.AllOf(
			cetestv2.HasType(event.String()),
		)
	}

	testCases := map[string]struct {
		task        func(namespace string) *v1beta1.Task
		matcherSets []AssertionSet
	}{
		"single-task": {
			task: func(namespace string) *v1beta1.Task {
				return &v1beta1.Task{
					ObjectMeta: metav1.ObjectMeta{Name: "single-task", Namespace: namespace},
					Spec: v1beta1.TaskSpec{
						// This was the digest of the latest tag as of 8/12/2019
						Steps: []v1beta1.Step{
							successStep,
						},
					},
				}
			},
			matcherSets: []AssertionSet{
				{N: 1, MatcherGen: assertSetGetFunc(resources.CloudEventTypeStepStarted)},
				{N: 1, MatcherGen: assertSetGetFunc(resources.CloudEventTypeStepSucceeded)},
			},
		},
		"double-task": {
			task: func(namespace string) *v1beta1.Task {
				return &v1beta1.Task{
					ObjectMeta: metav1.ObjectMeta{Name: "double-task", Namespace: namespace},
					Spec: v1beta1.TaskSpec{
						// This was the digest of the latest tag as of 8/12/2019
						Steps: []v1beta1.Step{
							successStep, successStep,
						},
					},
				}
			},
			matcherSets: []AssertionSet{
				{N: 1, MatcherGen: assertSetGetFunc(resources.CloudEventTypeStepStarted)},
				{N: 1, MatcherGen: assertSetGetFunc(resources.CloudEventTypeStepSucceeded)},
				{N: 1, MatcherGen: assertSetGetFunc(resources.CloudEventTypeStepStarted)},
				{N: 1, MatcherGen: assertSetGetFunc(resources.CloudEventTypeStepSucceeded)},
			},
		},
		"double-task-fail": {
			task: func(namespace string) *v1beta1.Task {
				return &v1beta1.Task{
					ObjectMeta: metav1.ObjectMeta{Name: "double-task-fail", Namespace: namespace},
					Spec: v1beta1.TaskSpec{
						// This was the digest of the latest tag as of 8/12/2019
						Steps: []v1beta1.Step{
							successStep, failStep,
						},
					},
				}
			},
			matcherSets: []AssertionSet{
				{N: 1, MatcherGen: assertSetGetFunc(resources.CloudEventTypeStepStarted)},
				{N: 1, MatcherGen: assertSetGetFunc(resources.CloudEventTypeStepSucceeded)},
				{N: 1, MatcherGen: assertSetGetFunc(resources.CloudEventTypeStepStarted)},
				{N: 1, MatcherGen: assertSetGetFunc(resources.CloudEventTypeStepFailed)},
			},
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			EventAssertion(t, test.task, test.matcherSets)
		})
	}
}
