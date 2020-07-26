package main

import (
	// This defines the shared main for injected controllers.
	"knative.dev/pkg/injection/sharedmain"

	"github.com/tom24d/step-observe-controller/pkg/reconciler"
)

const (
	component = "step-observe-reconciler"
)

func main() {
	sharedmain.Main(component, reconciler.NewController)
	// TODO add logic to watch config-defaults and set default-sink as CloudEvent sink to ctx or something
}
