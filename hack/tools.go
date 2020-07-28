// +build tools

package tools

import (
	_ "github.com/tektoncd/plumbing"
	_ "github.com/tektoncd/plumbing/scripts"

	// TODO remove this
	// for e2e test
	_ "knative.dev/eventing/test/test_images/recordevents"
	_ "knative.dev/eventing/config/core/configmaps"
)
