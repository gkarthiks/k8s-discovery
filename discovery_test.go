package discovery

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	"k8s.io/client-go/kubernetes/fake"
)

func newTestSimpleK8s() *K8s {
	client := K8s{}
	client.Clientset = fake.NewSimpleClientset()
	return &client
}

func TestGetVersionDefault(t *testing.T) {
	k8s := newTestSimpleK8s()
	v, err := k8s.GetVersion()
	if err != nil {
		t.Fatal("getVersion should not raise an error")
	}
	expected := "v0.0.0-master+$Format:%h$"
	if v != expected {
		t.Fatal("getVersion should return " + expected)
	}
}

func TestGetNamespace(t *testing.T) {
	k8s := newTestSimpleK8s()
	os.Setenv("POD_NAMESPACE", "default")
	ns, err := k8s.GetNamespace()
	if err != nil {
		t.Error("Expected result is default")
	}
	assert.Equal(t, "default", ns)
}
