package step

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"

	"knative.dev/pkg/logging"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/tektoncd/pipeline/pkg/apis/config"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	tektoncloudevent "github.com/tektoncd/pipeline/pkg/reconciler/events/cloudevent"
)

const (
	CloudEventSource = "/api/v1/namespaces/tekton-pipelines/pods/tekton-step-observer"
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
	PodRef    *corev1.ObjectReference `json:"podRef,omitempty"`
	Log       string                  `json:"log,omitempty"` // TODO evaluate the security risk.
	Step      *v1beta1.Step           `json:"step,omitempty"`
	StepState *v1beta1.StepState      `json:"stepState,omitempty"`
}

func (d *TektonStepCloudEvent) Emit(ctx context.Context, eventType TektonPluginEventType) {
	logger := logging.FromContext(ctx)
	configs := config.FromContextOrDefaults(ctx)
	sendCloudEvents := (configs.Defaults.DefaultCloudEventsSink != "")
	if !sendCloudEvents {
		return
	}
	ctx = cloudevents.ContextWithTarget(ctx, configs.Defaults.DefaultCloudEventsSink)
	ctx = cloudevents.ContextWithRetriesExponentialBackoff(ctx, 10*time.Millisecond, 10)
	cli := tektoncloudevent.Get(ctx)

	event := cloudevents.NewEvent()
	event.SetType(eventType.String())
	event.SetSource(CloudEventSource)

	if tm, err := GetEventTime(d.StepState, eventType); err != nil {
		logger.Errorf("failed to get time: %v", err)
	} else {
		event.SetTime(*tm)
	}

	err := event.SetData(cloudevents.ApplicationJSON, d)
	if err != nil {
		logger.Errorf("failed to marshal payload :%v", err)
	}

	if result := cli.Send(ctx, event); !cloudevents.IsACK(result) {
		logger.Errorf("failed to send CloudEvent: %v", result)
	} else {
		logger.Infof("Step:%s for %v has been sent successfully.", d.StepState.Name, eventType)
	}
}
