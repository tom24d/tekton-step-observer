package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"knative.dev/pkg/logging"
)

type CloudEventSent struct {
	Steps []Reported
}

func (s CloudEventSent) MarshalString() (string, error) {
	data, err := json.Marshal(s)
	return string(data), err
}

func UnmarshalString(str string, reported *CloudEventSent) error {
	return json.Unmarshal([]byte(str), reported)
}

func (s *CloudEventSent) MarkEvent(ctx context.Context, stepName string, eventType CloudEventType) (*CloudEventSent, error) {
	logger := logging.FromContext(ctx)
	logger.Infof("MarkEvent() for %s", stepName)
	logger.Infof("EventType: %s", eventType)
	logger.Infof("EventType: %v", eventType)

	for i, step := range s.Steps {
		if step.StepName == stepName {
			s.Steps[i].ReportedType = append(step.ReportedType, eventType)
			logger.Infof("MarkEvent() marked as %v", s)
			return s, nil
		}
	}

	return nil, fmt.Errorf("step: %s not found. could not mark event: %v to %v", stepName, eventType, s.Steps)
}

func (s *CloudEventSent) IsMarked(stepName string, eventType CloudEventType) bool {
	for _, step := range s.Steps {
		if step.StepName == stepName {
			return step.HasCloudEventType(eventType)
		}
	}
	return false // When no stepName step contained
}

func (s *CloudEventSent) GetReported(stepName string) (*Reported, error) {
	for _, step := range s.Steps {
		if step.StepName == stepName {
			return &step, nil
		}
	}
	return nil, fmt.Errorf("no reported struct contained")
}

type Reported struct {
	StepName     string           `json:"stepName"`
	ReportedType []CloudEventType `json:"reportedType"`
}

func (r *Reported) HasCloudEventType(eventType CloudEventType) bool {
	for _, s := range r.ReportedType {
		if s == eventType {
			return true
		}
	}
	return false
}
