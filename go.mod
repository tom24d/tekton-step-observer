module github.com/tom24d/step-observe-controller

go 1.14

require (
	github.com/cloudevents/sdk-go/v2 v2.2.0
	github.com/google/go-cmp v0.4.1
	github.com/hashicorp/go-multierror v1.1.0
	github.com/tektoncd/pipeline v0.15.1-0.20200827134110-9be0bb349825
	github.com/tektoncd/plumbing v0.0.0-20200828211711-efd006b76cb1
	go.uber.org/zap v1.15.0
	k8s.io/api v0.17.6
	k8s.io/apimachinery v0.17.6
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/eventing v0.16.2-0.20200827182806-c6f3501d66dc
	knative.dev/pkg v0.0.0-20200702222342-ea4d6e985ba0
)

replace k8s.io/client-go => k8s.io/client-go v0.17.6

replace k8s.io/apimachinery => k8s.io/apimachinery v0.17.6

replace k8s.io/api => k8s.io/api v0.17.6

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.6

replace k8s.io/apiserver => k8s.io/apiserver v0.17.6
