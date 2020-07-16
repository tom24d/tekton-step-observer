package resources

type CloudEventType string

const (
	CloudEventTypeStepStarted   CloudEventType = "dev.tekton.event.plugin.step.started.v1"
	CloudEventTypeStepFailed    CloudEventType = "dev.tekton.event.plugin.step.failed.v1"
	CloudEventTypeStepSucceeded CloudEventType = "dev.tekton.event.plugin.step.succeeded.v1"
	CloudEventTypeStepSkipped   CloudEventType = "dev.tekton.event.plugin.step.skipped.v1"
)
