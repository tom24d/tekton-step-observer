package test

import (
	"testing"

	eventingtest "knative.dev/eventing/test/lib"

	pipelinetest "github.com/tektoncd/pipeline/test"
)

type ClientSet struct {
	pipelineTestClient clients
	eventingTestClient eventingtest.
}

func Setup(t *testing.T, fn ...func(*testing.T, *ClientSet, string)) (ClientSet, error) {

}