package test

import (
	"os"
	"testing"

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
