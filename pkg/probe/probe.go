package probe

import (
	"fmt"
	"os"
	"strings"

	"github.com/rancher/rancherd/pkg/config"
	"github.com/rancher/rancherd/pkg/roles"
	"github.com/rancher/rancherd/pkg/self"
	"github.com/rancher/system-agent/pkg/applyinator"
	"github.com/rancher/system-agent/pkg/prober"
)

var probes = map[string]prober.Probe{
	"kube-apiserver": {
		InitialDelaySeconds: 1,
		TimeoutSeconds:      5,
		SuccessThreshold:    1,
		FailureThreshold:    2,
		HTTPGetAction: prober.HTTPGetAction{
			URL:        "https://127.0.0.1:6443/readyz",
			CACert:     "/var/lib/rancher/%s/server/tls/server-ca.crt",
			ClientCert: "/var/lib/rancher/%s/server/tls/client-kube-apiserver.crt",
			ClientKey:  "/var/lib/rancher/%s/server/tls/client-kube-apiserver.key",
		},
	},
	"kube-scheduler": {
		InitialDelaySeconds: 1,
		TimeoutSeconds:      5,
		SuccessThreshold:    1,
		FailureThreshold:    2,
		HTTPGetAction: prober.HTTPGetAction{
			URL: "http://127.0.0.1:10251/healthz",
		},
	},
	"kube-controller-manager": {
		InitialDelaySeconds: 1,
		TimeoutSeconds:      5,
		SuccessThreshold:    1,
		FailureThreshold:    2,
		HTTPGetAction: prober.HTTPGetAction{
			URL: "http://127.0.0.1:10252/healthz",
		},
	},
	"kubelet": {
		InitialDelaySeconds: 1,
		TimeoutSeconds:      5,
		SuccessThreshold:    1,
		FailureThreshold:    2,
		HTTPGetAction: prober.HTTPGetAction{
			URL: "http://127.0.0.1:10248/healthz",
		},
	},
}

func replaceRuntime(str string, runtime config.Runtime) string {
	if !strings.Contains(str, "%s") {
		return str
	}
	return fmt.Sprintf(str, runtime)
}

func ProbesForRole(config *config.RuntimeConfig, runtime config.Runtime) map[string]prober.Probe {
	if roles.IsControlPlane(config.Role) {
		return AllProbes(runtime)
	}
	return replaceRuntimeForProbes(map[string]prober.Probe{
		"kubelet": probes["kubelet"],
	}, runtime)
}

func AllProbes(runtime config.Runtime) map[string]prober.Probe {
	return replaceRuntimeForProbes(probes, runtime)
}

func replaceRuntimeForProbes(probes map[string]prober.Probe, runtime config.Runtime) map[string]prober.Probe {
	result := map[string]prober.Probe{}
	for k, v := range probes {
		v.HTTPGetAction.CACert = replaceRuntime(v.HTTPGetAction.CACert, runtime)
		v.HTTPGetAction.ClientCert = replaceRuntime(v.HTTPGetAction.ClientCert, runtime)
		v.HTTPGetAction.ClientKey = replaceRuntime(v.HTTPGetAction.ClientKey, runtime)
		result[k] = v
	}
	return result
}

func ToInstruction(imageOverride string, systemDefaultRegistry string, k8sVersion string) (*applyinator.Instruction, error) {
	cmd, err := self.Self()
	if err != nil {
		return nil, fmt.Errorf("resolving location of %s: %w", os.Args[0], err)
	}
	return &applyinator.Instruction{
		Name:       "probes",
		SaveOutput: true,
		Args:       []string{"probe"},
		Command:    cmd,
	}, nil
}
