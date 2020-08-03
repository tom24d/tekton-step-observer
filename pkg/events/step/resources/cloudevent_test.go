package resources

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"

	rtesting "knative.dev/pkg/reconciler/testing"

	"github.com/tektoncd/pipeline/pkg/apis/config"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	tektoncloudevent "github.com/tektoncd/pipeline/pkg/reconciler/events/cloudevent"
)

func TestEmit(t *testing.T) {
	eventData := TektonStepCloudEvent{
		PodRef: &corev1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Pod",
			Name:       "test-name",
			Namespace:  "test-namespace",
		},
		Log:       "this is log",
		Step:      &v1beta1.Step{},
		StepState: &v1beta1.StepState{},
	}

	eventTypes := []TektonPluginEventType{
		CloudEventTypeStepStarted,
		CloudEventTypeStepFailed,
		CloudEventTypeStepSucceeded,
		CloudEventTypeStepSkipped,
	}

	testcases := []struct {
		name           string
		data           map[string]string
		wantEvent      string
		wantCloudEvent bool
	}{{
		name:           "without sink",
		data:           map[string]string{},
		wantEvent:      "",
		wantCloudEvent: false,
	}, {
		name:           "with empty string sink",
		data:           map[string]string{"default-cloud-events-sink": ""},
		wantEvent:      "",
		wantCloudEvent: false,
	}, {
		name:           "with sink",
		data:           map[string]string{"default-cloud-events-sink": "http://mysink"},
		wantEvent:      "Normal Started",
		wantCloudEvent: true,
	}}

	for _, et := range eventTypes {
		for _, tc := range testcases {
			t.Run(tc.name+"/"+et.String(), func(t *testing.T) {
				// Setup the context and seed test data
				ctx, _ := rtesting.SetupFakeContext(t)
				ctx = tektoncloudevent.WithClient(ctx, &tektoncloudevent.FakeClientBehaviour{SendSuccessfully: true})
				fakeClient := tektoncloudevent.Get(ctx).(tektoncloudevent.FakeClient)

				// Setup the config and add it to the context
				defaults, _ := config.NewDefaultsFromMap(tc.data)
				featureFlags, _ := config.NewFeatureFlagsFromMap(map[string]string{})
				cfg := &config.Config{
					Defaults:     defaults,
					FeatureFlags: featureFlags,
				}
				ctx = config.ToContext(ctx, cfg)

				go eventData.Emit(ctx, et)
				if tc.wantCloudEvent {
					if err := checkCloudEvents(t, &fakeClient, t.Name(), `(s?)`+et.String()); err != nil {
						t.Fatalf(err.Error())
					}
				}
			})
		}
	}
}

func eventFromChannel(c chan string, testName string, wantEvent string) error {
	timer := time.NewTimer(1 * time.Second)
	select {
	case event := <-c:
		if wantEvent == "" {
			return fmt.Errorf("received event \"%s\" for %s but none expected", event, testName)
		}
		matching, err := regexp.MatchString(wantEvent, event)
		if err == nil {
			if !matching {
				return fmt.Errorf("expected event \"%s\" but got \"%s\" instead for %s", wantEvent, event, testName)
			}
		}
	case <-timer.C:
		if wantEvent != "" {
			return fmt.Errorf("received no events for %s but %s expected", testName, wantEvent)
		}
	}
	return nil
}

func checkCloudEvents(t *testing.T, fce *tektoncloudevent.FakeClient, testName string, wantEvent string) error {
	t.Helper()
	return eventFromChannel(fce.Events, testName, wantEvent)
}
