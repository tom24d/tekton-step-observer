package step

import (
	"fmt"
)

type EmissionStatuses struct {
	Statuses []EmissionStatus `json:"statuses"`
}

func (s *EmissionStatuses) IsMarked(stepName string, eventType TektonPluginEventType) bool {
	for i, step := range s.Statuses {
		if step.Name == stepName {
			return s.Statuses[i].IsMarked(eventType)
		}
	}
	return false
}

func (s *EmissionStatuses) GetStatus(stepName string) (*EmissionStatus, error) {
	for i, step := range s.Statuses {
		if step.Name == stepName {
			return &s.Statuses[i], nil
		}
	}
	return nil, fmt.Errorf("no reported struct contained")
}

type EmissionStatus struct {
	Name    string                  `json:"name"`
	Emitted []TektonPluginEventType `json:"emitted"`
}

func (r *EmissionStatus) MarkEvent(eventType TektonPluginEventType) error {
	if r.IsMarked(eventType) {
		return fmt.Errorf("%v is already marked in the step:%s", eventType, r.Name)
	}
	r.Emitted = append(r.Emitted, eventType)
	return nil
}

func (r *EmissionStatus) IsMarked(eventType TektonPluginEventType) bool {
	for _, s := range r.Emitted {
		if s == eventType {
			return true
		}
	}
	return false
}
