module github.com/tom24d/step-observe-controller

go 1.14

require (
	github.com/tektoncd/pipeline v0.11.0
	github.com/tektoncd/plumbing v0.0.0-20200710074257-b4976a02c01c
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/pkg v0.0.0-20200710003319-43f4f824e3a3
)

replace k8s.io/client-go => k8s.io/client-go v0.17.6

replace k8s.io/apimachinery => k8s.io/apimachinery v0.17.6

replace k8s.io/api => k8s.io/api v0.17.6
