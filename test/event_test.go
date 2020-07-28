// +build e2e

package test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventingtestlib "knative.dev/eventing/test/lib"
	"knative.dev/eventing/test/lib/recordevents"

	knativetest "knative.dev/pkg/test"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	pipelineTest "github.com/tektoncd/pipeline/test"

	"github.com/tom24d/step-observe-controller/pkg/events/step/resources"

	cetestv2 "github.com/cloudevents/sdk-go/v2/test"
)

func Test_Event(t *testing.T) {

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
	task := &v1beta1.Task{
		ObjectMeta: metav1.ObjectMeta{Name: "status-task", Namespace: client.Namespace},
		Spec: v1beta1.TaskSpec{
			// This was the digest of the latest tag as of 8/12/2019
			Steps: []v1beta1.Step{
				v1beta1.Step{Container: corev1.Container{
					Image:   "busybox@sha256:895ab622e92e18d6b461d671081757af7dbaa3b00e3e28e12505af7817f73649",
					Command: []string{"/bin/sh"},
					Args:    []string{"-c", "echo hello1"},
				},
				},
			},
		},
	}

	if _, err := pipelineClient.TaskClient.Create(task); err != nil {
		t.Fatalf("Failed to create Task: %s", err)
	}

	taskRun := &v1beta1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{Name: taskRunName, Namespace: client.Namespace},
		Spec: v1beta1.TaskRunSpec{
			TaskRef: &v1beta1.TaskRef{Name: "status-task"},
		},
	}

	if _, err := pipelineClient.TaskRunClient.Create(taskRun); err != nil {
		t.Fatalf("Failed to create TaskRun: %s", err)
	}

	if err := WaitForTaskRunState(pipelineClient, taskRunName, pipelineTest.TaskRunSucceed(taskRunName)); err != nil {
		t.Fatalf("Failed to wait TaskRun: %s", err)
	}
	t.Logf("Asserting CloudEvent in recordevents pod...")

	// multi-assert event
	eventTracker.AssertExact(1, recordevents.MatchEvent(cetestv2.AllOf(cetestv2.HasType(resources.CloudEventTypeStepStarted.String()))))
	eventTracker.AssertExact(1, recordevents.MatchEvent(cetestv2.AllOf(cetestv2.HasType(resources.CloudEventTypeStepSucceeded.String()))))
}
