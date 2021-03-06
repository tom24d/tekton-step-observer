//+build e2e

package test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cetestv2 "github.com/cloudevents/sdk-go/v2/test"

	eventinghelpers "knative.dev/eventing/test/e2e/helpers"
	eventingtestlib "knative.dev/eventing/test/lib"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	"github.com/tom24d/step-observe-controller/pkg/events/step"
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

	assertSetGetFunc := func(event step.TektonPluginEventType, n int) AssertionSet {
		return AssertionSet{
			N: n,
			Matchers: []cetestv2.EventMatcher{
				cetestv2.HasType(event.String()),
				cetestv2.HasSource(step.CloudEventSource),
			},
			eventType: event,
		}
	}

	testCases := map[string]struct {
		task        func(namespace string) *v1beta1.Task
		matcherSets []AssertionSet
	}{
		"success": {
			task: func(namespace string) *v1beta1.Task {
				return &v1beta1.Task{
					ObjectMeta: metav1.ObjectMeta{Name: "success", Namespace: namespace},
					Spec: v1beta1.TaskSpec{
						// This was the digest of the latest tag as of 8/12/2019
						Steps: []v1beta1.Step{
							successStep,
						},
					},
				}
			},
			matcherSets: []AssertionSet{
				assertSetGetFunc(step.CloudEventTypeStepStarted, 1),
				assertSetGetFunc(step.CloudEventTypeStepSucceeded, 1),
			},
		},
		"success-success": {
			task: func(namespace string) *v1beta1.Task {
				return &v1beta1.Task{
					ObjectMeta: metav1.ObjectMeta{Name: "success-success", Namespace: namespace},
					Spec: v1beta1.TaskSpec{
						// This was the digest of the latest tag as of 8/12/2019
						Steps: []v1beta1.Step{
							successStep, successStep,
						},
					},
				}
			},
			matcherSets: []AssertionSet{
				assertSetGetFunc(step.CloudEventTypeStepStarted, 2),
				assertSetGetFunc(step.CloudEventTypeStepSucceeded, 2),
			},
		},
		"success-fail": {
			task: func(namespace string) *v1beta1.Task {
				return &v1beta1.Task{
					ObjectMeta: metav1.ObjectMeta{Name: "success-fail", Namespace: namespace},
					Spec: v1beta1.TaskSpec{
						// This was the digest of the latest tag as of 8/12/2019
						Steps: []v1beta1.Step{
							successStep, failStep,
						},
					},
				}
			},
			matcherSets: []AssertionSet{
				assertSetGetFunc(step.CloudEventTypeStepStarted, 2),
				assertSetGetFunc(step.CloudEventTypeStepSucceeded, 1),
				assertSetGetFunc(step.CloudEventTypeStepFailed, 1),
			},
		},
		"success-fail-skip-skip": {
			task: func(namespace string) *v1beta1.Task {
				return &v1beta1.Task{
					ObjectMeta: metav1.ObjectMeta{Name: "success-fail-skip-skip", Namespace: namespace},
					Spec: v1beta1.TaskSpec{
						// This was the digest of the latest tag as of 8/12/2019
						Steps: []v1beta1.Step{
							successStep, failStep, successStep, failStep,
						},
					},
				}
			},
			matcherSets: []AssertionSet{
				assertSetGetFunc(step.CloudEventTypeStepStarted, 2),
				assertSetGetFunc(step.CloudEventTypeStepSucceeded, 1),
				assertSetGetFunc(step.CloudEventTypeStepFailed, 1),
				assertSetGetFunc(step.CloudEventTypeStepSkipped, 2),
			},
		},
		"success-25": {
			task: func(namespace string) *v1beta1.Task {
				return &v1beta1.Task{
					ObjectMeta: metav1.ObjectMeta{Name: "success-25", Namespace: namespace},
					Spec: v1beta1.TaskSpec{
						// This was the digest of the latest tag as of 8/12/2019
						Steps: func() []v1beta1.Step {
							steps := make([]v1beta1.Step, 25)
							for i := range steps {
								steps[i] = successStep
							}
							return steps
						}(),
					},
				}
			},
			matcherSets: []AssertionSet{
				assertSetGetFunc(step.CloudEventTypeStepStarted, 25),
				assertSetGetFunc(step.CloudEventTypeStepSucceeded, 25),
			},
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			RunTests(&channelTestRunner, t, eventingtestlib.FeatureBasic, func(st *testing.T, component metav1.TypeMeta) {
				brokerCreator := eventinghelpers.ChannelBasedBrokerCreator(component, brokerClass)
				EventAssertion(st, test.task, test.matcherSets, brokerCreator)
			})
		})
	}
}
