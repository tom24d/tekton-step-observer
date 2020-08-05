package test

import (
	"encoding/json"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	pkgTest "knative.dev/pkg/test"
)

func PatchDefaultCloudEventSinkOrFail(t *testing.T, c *pkgTest.KubeClient, sink string, ns string) {

	cm := corev1.ConfigMap{
		Data: map[string]string{
			"default-cloud-events-sink": sink,
		},
	}
	data, err := json.Marshal(cm)
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.Kube.CoreV1().ConfigMaps("tekton-pipelines").Patch("config-defaults", types.MergePatchType, data)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("\"default-cloud-events-sink\" is set to: %s", sink)
}
