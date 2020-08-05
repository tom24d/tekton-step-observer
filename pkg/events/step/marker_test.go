package step

import (
	"testing"

	"github.com/tektoncd/pipeline/test/diff"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshalString(t *testing.T) {

}

func TestEmissionStatuses_MarshalString(t *testing.T) {
}

func TestEmissionStatuses_mark_check(t *testing.T) {
	const name = "hello1"
	const marking = CloudEventTypeStepFailed
	s := EmissionStatus{
		Name:    name,
		Emitted: []TektonPluginEventType{CloudEventTypeStepStarted},
	}
	statuses := &EmissionStatuses{
		Statuses: []EmissionStatus{s},
	}

	gotStatus, err := statuses.GetStatus(name)
	if err != nil {
		t.Fatal(err)
	}
	err = gotStatus.MarkEvent(marking)
	if err != nil {
		t.Fatal(err)
	}
	if !statuses.IsMarked(name, marking) {
		t.Fail()
	}
}

func TestEmissionStatuses_IsMarked(t *testing.T) {
	const name = "hello1"
	s := EmissionStatus{
		Name:    name,
		Emitted: []TektonPluginEventType{CloudEventTypeStepStarted},
	}
	statuses := EmissionStatuses{
		Statuses: []EmissionStatus{s},
	}

	if !statuses.IsMarked(name, CloudEventTypeStepStarted) {
		t.Fail()
	}
	if statuses.IsMarked(name, CloudEventTypeStepSucceeded) {
		t.Fail()
	}
	if statuses.IsMarked("fake", CloudEventTypeStepStarted) {
		t.Fail()
	}
}

func TestEmissionStatuses_GetStatus(t *testing.T) {
	const name = "hello1"
	want := EmissionStatus{
		Name:    name,
		Emitted: []TektonPluginEventType{CloudEventTypeStepStarted},
	}
	statuses := EmissionStatuses{
		Statuses: []EmissionStatus{want,
			{Name: "hello2", Emitted: []TektonPluginEventType{CloudEventTypeStepSkipped}}},
	}
	got, err := statuses.GetStatus(name)
	if err != nil {
		t.Fatal(err)
	}
	if d := cmp.Diff(want, got); d != "" {
		diff.PrintWantGot(d)
	}
}

func TestEmissionStatuses_GetStatus_Err(t *testing.T) {
	const name = "hello1"
	s := EmissionStatuses{
		Statuses: []EmissionStatus{
			{Name: name, Emitted: []TektonPluginEventType{CloudEventTypeStepSkipped}}},
	}

	got, err := s.GetStatus("fake")
	if err == nil {
		t.Fatalf("error expected, got=%v", got)
	}
}

func TestEmissionStatus_MarkEvent(t *testing.T) {
	const name = "hello1"
	want := EmissionStatus{
		Name:    name,
		Emitted: []TektonPluginEventType{CloudEventTypeStepStarted},
	}
	got := EmissionStatus{
		Name: name,
	}
	err := got.MarkEvent(CloudEventTypeStepStarted)
	if err != nil {
		t.Fatal(err)
	}
	if d := cmp.Diff(want, got); d != "" {
		t.Fatalf(diff.PrintWantGot(d))
	}
}

func TestEmissionStatus_MarkEvent_Err(t *testing.T) {
	const name = "hello1"
	s := EmissionStatus{
		Name:    name,
		Emitted: []TektonPluginEventType{CloudEventTypeStepStarted},
	}
	err := s.MarkEvent(CloudEventTypeStepStarted)
	if err == nil {
		t.Fail()
	}
}

func TestEmissionStatus_IsMarked(t *testing.T) {
	obj := EmissionStatus{
		Name:    "hello1",
		Emitted: []TektonPluginEventType{CloudEventTypeStepStarted, CloudEventTypeStepFailed},
	}
	if !obj.IsMarked(CloudEventTypeStepStarted) {
		t.Fail()
	}
	if obj.IsMarked(CloudEventTypeStepSkipped) {
		t.Fail()
	}
}
