package test

import (
	"fmt"
	"os"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/eventing/test"
	testlib "knative.dev/eventing/test/lib"
)

var channelTestRunner testlib.ComponentsTestRunner
var brokerClass string

func TestMain(m *testing.M) {
	test.InitializeEventingFlags()
	channelTestRunner = testlib.ComponentsTestRunner{
		ComponentFeatureMap: testlib.ChannelFeatureMap,
		ComponentsToTest:    test.EventingFlags.Channels,
	}
	brokerClass = test.EventingFlags.BrokerClass

	os.Exit(m.Run())
}

// copy from eventing
func RunTests(
	tr *testlib.ComponentsTestRunner,
	t *testing.T,
	feature testlib.Feature,
	testFunc func(st *testing.T, component metav1.TypeMeta),
) {
	for _, component := range tr.ComponentsToTest {
		// If a component is not present in the map, then assume it has all properties. This is so an
		// unknown component (e.g. a Channel) can be specified via a dedicated flag (e.g. --channels) and have tests run.
		// TODO Use a flag to specify the features of the flag based component, rather than assuming
		// it supports all features.
		features, present := tr.ComponentFeatureMap[component]
		if !present || contains(features, feature) {
			t.Run(fmt.Sprintf("%s-%s", component.Kind, component.APIVersion), func(st *testing.T) {
				testFunc(st, component)
			})
		}
	}
}

// copy from eventing
func contains(features []testlib.Feature, feature testlib.Feature) bool {
	for _, f := range features {
		if f == feature {
			return true
		}
	}
	return false
}