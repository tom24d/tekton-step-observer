package test

import (
	"knative.dev/pkg/apis"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventingtestlib "knative.dev/eventing/test/lib"
	"knative.dev/eventing/test/lib/recordevents"

	knativetest "knative.dev/pkg/test"

	cetestv2 "github.com/cloudevents/sdk-go/v2/test"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

type AssertionSet struct {
	N          int
	MatcherGen cetestv2.EventMatcher
}

func EventAssertion(t *testing.T, task func(namespace string) *v1beta1.Task, assertionSet []AssertionSet) {

	t.Helper()

	const (
		recordEventPodName = "e2e-step-observer-logger-event-tracker"
		taskRunName        = "e2e-test-step-observed-run"
	)

	client := eventingtestlib.Setup(t, false)
	defer eventingtestlib.TearDown(client)

	pipelineClient := newClients(t, knativetest.Flags.Kubeconfig, knativetest.Flags.Cluster, client.Namespace)

	// create event logger eventSender and service
	eventTracker, ePod := recordevents.StartEventRecordOrFail(client, recordEventPodName)
	defer eventTracker.Cleanup()

	// set default-sink
	PatchDefaultCloudEventSinkOrFail(t, client.Kube, "http://"+client.GetServiceHost(ePod.Name), client.Namespace)

	t.Logf("Creating Task and TaskRun in namespace %s", client.Namespace)

	if _, err := pipelineClient.TaskClient.Create(task(client.Namespace)); err != nil {
		t.Fatalf("Failed to create Task: %s", err)
	}

	taskRun := &v1beta1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{Name: taskRunName, Namespace: client.Namespace},
		Spec: v1beta1.TaskRunSpec{
			TaskRef: &v1beta1.TaskRef{Name: task(client.Namespace).Name},
		},
	}

	if _, err := pipelineClient.TaskRunClient.Create(taskRun); err != nil {
		t.Fatalf("Failed to create TaskRun: %s", err)
	}

	if err := WaitForTaskRunState(pipelineClient, taskRunName, func(ca apis.ConditionAccessor) (bool, error) {
		c := ca.GetCondition(apis.ConditionSucceeded)
		if c != nil {
			return true, nil
		}
		return false, nil
	}); err != nil {
		t.Fatalf("Failed to wait TaskRun: %s", err)
	}
	t.Log("Asserting CloudEvent...")

	// multi-assert event
	for _, s := range assertionSet {
		eventTracker.AssertExact(s.N, recordevents.MatchEvent(s.MatcherGen))
	}
}
