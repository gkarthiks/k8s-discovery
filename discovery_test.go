package discovery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewK8s(t *testing.T) {
	const validConfig = `apiVersion: v1
kind: Config
clusters:
- name: test
  cluster:
    server: https://127.0.0.1:6443
contexts:
- name: test
  context:
    cluster: test
    user: test
current-context: test
users:
- name: test
  user:
    token: test-token
`
	tests := []struct {
		name      string
		config    string
		missing   bool
		wantError string
	}{
		{name: "valid kubeconfig", config: validConfig},
		{name: "missing kubeconfig", missing: true, wantError: "no such file"},
		{name: "malformed kubeconfig", config: "clusters: [", wantError: "error loading config file"},
		{
			name: "invalid CA certificate",
			config: strings.Replace(validConfig, "    server:",
				"    certificate-authority-data: bm90IGEgY2VydGlmaWNhdGU=\n    server:", 1),
			wantError: "create Kubernetes client:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Force local configuration without using the developer's kubeconfig.
			t.Setenv("KUBERNETES_SERVICE_HOST", "")
			t.Setenv("KUBERNETES_SERVICE_PORT", "")
			configPath := filepath.Join(t.TempDir(), "config")
			t.Setenv("KUBECONFIG", configPath)
			if !tt.missing {
				require.NoError(t, os.WriteFile(configPath, []byte(tt.config), 0600))
			}

			client, err := NewK8s()
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				assert.Nil(t, client)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, client)
			assert.NotNil(t, client.Clientset)
			assert.NotNil(t, client.MetricsClientSet)
			require.NotNil(t, client.RestConfig)
			assert.Equal(t, "https://127.0.0.1:6443", client.RestConfig.Host)
		})
	}
}

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
	expected := "v0.0.0-master+$Format:%H$"
	if v != expected {
		t.Fatal("getVersion should return " + expected)
	}
}

func TestGetNamespace(t *testing.T) {
	k8s := newTestSimpleK8s()
	t.Setenv("POD_NAMESPACE", "default")
	ns, err := k8s.GetNamespace()
	if err != nil {
		t.Error("Expected result is default")
	}
	assert.Equal(t, "default", ns)
}
