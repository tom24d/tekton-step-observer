module github.com/tom24d/step-observe-controller

go 1.14

require (
	github.com/tektoncd/plumbing v0.0.0-20200710074257-b4976a02c01c
	knative.dev/pkg v0.0.0-20200710003319-43f4f824e3a3
)

replace k8s.io/client-go => k8s.io/client-go v0.17.6

replace k8s.io/apimachinery => k8s.io/apimachinery v0.17.6

replace k8s.io/api => k8s.io/api v0.17.6
