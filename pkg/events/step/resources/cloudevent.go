package resources

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	tektoncloudevent "github.com/tektoncd/pipeline/pkg/reconciler/events/cloudevent"
)

type TektonPluginEventType string

const (
	CloudEventTypeStepStarted   TektonPluginEventType = "dev.tekton.events.plugin.step.started.v1"
	CloudEventTypeStepFailed    TektonPluginEventType = "dev.tekton.events.plugin.step.failed.v1"
	CloudEventTypeStepSucceeded TektonPluginEventType = "dev.tekton.events.plugin.step.succeeded.v1"
	CloudEventTypeStepSkipped   TektonPluginEventType = "dev.tekton.events.plugin.step.skipped.v1"
)

func (c TektonPluginEventType) String() string {
	return string(c)
}

type TektonStepCloudEvent struct {
	Pod    *corev1.Pod        `json:"pod,omitempty"`
	Log    string             `json:"log,omitempty"`
	State  *v1beta1.StepState `json:"state,omitempty"`
	Detail *v1beta1.Step      `json:"detail,omitempty"`
}

func (d *TektonStepCloudEvent) Emit(ctx context.Context, eventType TektonPluginEventType) error {
	cli := tektoncloudevent.Get(ctx)

	event := cloudevents.NewEvent()
	event.SetType(eventType.String())
	event.SetSource("tbd")
	err := event.SetData(cloudevents.ApplicationJSON, d)
	if err != nil {
		return fmt.Errorf("failed to marshal payload :%v", err)
	}

	if result := cli.Send(ctx, event); cloudevents.IsUndelivered(result) {
		return fmt.Errorf("failed to send CloudEvent: %v", result)
	}
	return nil
}
