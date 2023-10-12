// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package e2e

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_test/e2e"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_test/validate"
	"github.com/steadybit/extension-kubernetes/extcluster"
	"github.com/steadybit/extension-kubernetes/extcontainer"
	"github.com/steadybit/extension-kubernetes/extdeployment"
	"github.com/steadybit/extension-kubernetes/extnode"
	"github.com/steadybit/extension-kubernetes/extpod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/strings/slices"
	"strings"
	"testing"
	"time"
)

func TestWithMinikube(t *testing.T) {
	extFactory := e2e.HelmExtensionFactory{
		Name: "extension-kubernetes",
		Port: 8088,
		ExtraArgs: func(m *e2e.Minikube) []string {
			return []string{
				"--set", "kubernetes.clusterName=e2e-cluster",
				"--set", "discovery.attributes.excludes.container={k8s.label.*}",
				"--set", "logging.level=debug",
			}
		},
	}

	e2e.WithDefaultMinikube(t, &extFactory, []e2e.WithMinikubeTestCase{
		{
			Name: "validate discovery",
			Test: validateDiscovery,
		},
		{
			Name: "discovery",
			Test: testDiscovery,
		},
		{
			Name: "checkRolloutReady",
			Test: testCheckRolloutReady,
		},
		{
			Name: "deletePod",
			Test: testDeletePod,
		},
		{
			Name: "drainNode",
			Test: testDrainNode,
		},
		{
			Name: "taintNode",
			Test: testTaintNode,
		},
		{
			Name: "scaleDeployment",
			Test: testScaleDeployment,
		},
	})
}

func validateDiscovery(t *testing.T, _ *e2e.Minikube, e *e2e.Extension) {
	assert.NoError(t, validate.ValidateEndpointReferences("/", e.Client))
}

func testCheckRolloutReady(t *testing.T, m *e2e.Minikube, e *e2e.Extension) {
	log.Info().Msg("Starting testCheckRolloutReady")

	nginxDeployment := e2e.NginxDeployment{Minikube: m}
	err := nginxDeployment.Deploy("nginx-check-rollout-ready")
	require.NoError(t, err, "failed to create deployment")
	defer func() { _ = nginxDeployment.Delete() }()

	tests := []struct {
		name            string
		wantedCompleted bool
	}{
		{
			name:            "should check status ok",
			wantedCompleted: true,
		},
		{
			name:            "should check status not completed",
			wantedCompleted: false,
		},
	}

	require.NoError(t, err)

	for _, tt := range tests {

		config := struct {
			Duration int `json:"duration"`
		}{
			Duration: 15000,
		}

		t.Run(tt.name, func(t *testing.T) {
			if tt.wantedCompleted {
				exec, err := m.PodExec(e.Pod, "extension", "kubectl", "rollout", "restart", "deployment/nginx-check-rollout-ready")
				require.NoError(t, err)
				log.Info().Msgf("kubectl rollout restart deployment/nginx-check-rollout-ready: %s", exec)
			} else {
				exec, err := m.PodExec(e.Pod, "extension", "kubectl", "rollout", "restart", "deployment/nginx-check-rollout-ready")
				require.NoError(t, err)
				log.Info().Msgf("kubectl rollout restart deployment/nginx-check-rollout-ready: %s", exec)
				exec, err = m.PodExec(e.Pod, "extension", "kubectl", "rollout", "pause", "deployment/nginx-check-rollout-ready")
				require.NoError(t, err)
				log.Info().Msgf("kubectl rollout pause deployment/nginx-check-rollout-ready: %s", exec)
			}

			target := action_kit_api.Target{
				Name: "test",
				Attributes: map[string][]string{
					"k8s.cluster-name": {"e2e-cluster"},
					"k8s.namespace":    {"default"},
					"k8s.deployment":   {"nginx-check-rollout-ready"},
				},
			}
			action, err := e.RunAction(extdeployment.RolloutStatusActionId, &target, config, nil)
			defer func() { _ = action.Cancel() }()
			require.NoError(t, err)

			err = action.Wait()
			if tt.wantedCompleted {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}

}

func testDiscovery(t *testing.T, m *e2e.Minikube, e *e2e.Extension) {
	log.Info().Msg("Starting testDiscovery")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	nginxDeployment := e2e.NginxDeployment{Minikube: m}
	err := nginxDeployment.Deploy("nginx")
	require.NoError(t, err, "failed to create deployment")
	defer func() { _ = nginxDeployment.Delete() }()

	target, err := e2e.PollForTarget(ctx, e, extdeployment.DeploymentTargetType, func(target discovery_kit_api.Target) bool {
		return e2e.HasAttribute(target, "k8s.deployment", "nginx")
	})

	require.NoError(t, err)
	assert.Equal(t, target.TargetType, extdeployment.DeploymentTargetType)
	assert.Equal(t, target.Attributes["k8s.namespace"][0], "default")
	assert.Equal(t, target.Attributes["k8s.deployment"][0], "nginx")
	assert.Equal(t, target.Attributes["k8s.deployment.label.app"][0], "nginx")
	assert.Equal(t, target.Attributes["k8s.cluster-name"][0], "e2e-cluster")
	assert.Contains(t, target.Attributes["k8s.pod.name"], nginxDeployment.Pods[0].Name)
	assert.Contains(t, target.Attributes["k8s.pod.name"], nginxDeployment.Pods[1].Name)
	assert.Equal(t, target.Attributes["k8s.distribution"][0], "kubernetes")

	enrichmentData, err := e2e.PollForEnrichmentData(ctx, e, extcontainer.KubernetesContainerEnrichmentDataType, func(enrichmentData discovery_kit_api.EnrichmentData) bool {
		return e2e.ContainsAttribute(enrichmentData.Attributes, "k8s.container.name", "nginx")
	})

	require.NoError(t, err)
	assert.Equal(t, enrichmentData.EnrichmentDataType, extcontainer.KubernetesContainerEnrichmentDataType)
	assert.Equal(t, enrichmentData.Attributes["k8s.container.name"][0], "nginx")
	assert.Equal(t, enrichmentData.Attributes["k8s.container.image"][0], "nginx:stable-alpine")
	assert.Equal(t, enrichmentData.Attributes["k8s.pod.label.app"][0], "nginx")
	assert.Equal(t, enrichmentData.Attributes["k8s.namespace"][0], "default")
	assert.Equal(t, enrichmentData.Attributes["k8s.node.name"][0], "e2e-docker")
	assert.NotContains(t, enrichmentData.Attributes, "k8s.label.app")

	podNames := make([]string, 0, len(nginxDeployment.Pods))
	for _, pod := range nginxDeployment.Pods {
		podNames = append(podNames, pod.Name)
	}
	assert.Contains(t, podNames, enrichmentData.Attributes["k8s.pod.name"][0])

	target, err = e2e.PollForTarget(ctx, e, extcluster.ClusterTargetType, func(target discovery_kit_api.Target) bool {
		return e2e.HasAttribute(target, "k8s.cluster-name", "e2e-cluster")
	})
	require.NoError(t, err)
	assert.Equal(t, target.TargetType, extcluster.ClusterTargetType)

	target, err = e2e.PollForTarget(ctx, e, extpod.PodTargetType, func(target discovery_kit_api.Target) bool {
		return e2e.HasAttribute(target, "k8s.deployment", "nginx")
	})
	require.NoError(t, err)
	assert.Equal(t, target.TargetType, extpod.PodTargetType)

	target, err = e2e.PollForTarget(ctx, e, extnode.NodeTargetType, func(target discovery_kit_api.Target) bool {
		return true
	})
	require.NoError(t, err)
	assert.Equal(t, target.TargetType, extnode.NodeTargetType)
	assert.Equal(t, "e2e-docker", target.Attributes["host.hostname"][0])
}

func testDeletePod(t *testing.T, m *e2e.Minikube, e *e2e.Extension) {
	log.Info().Msg("Starting testDeletePod")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Start Deployment with 2 pods
	nginxDeployment := e2e.NginxDeployment{Minikube: m}
	err := nginxDeployment.Deploy("nginx-test-delete-pod")
	require.NoError(t, err, "failed to create deployment")
	defer func() { _ = nginxDeployment.Delete() }()
	podName1 := nginxDeployment.Pods[0].Name
	podName2 := nginxDeployment.Pods[1].Name
	log.Info().Msgf("Pods before Attack: podName1: %v, podName2: %v", podName1, podName2)

	//Delete both pods
	_, err = e.RunAction(extpod.DeletePodActionId, &action_kit_api.Target{
		Name:       podName1,
		Attributes: map[string][]string{"k8s.pod.name": {podName1}, "k8s.namespace": {"default"}},
	}, nil, nil)
	require.NoError(t, err)
	_, err = e.RunAction(extpod.DeletePodActionId, &action_kit_api.Target{
		Name:       podName2,
		Attributes: map[string][]string{"k8s.pod.name": {podName2}, "k8s.namespace": {"default"}},
	}, nil, nil)
	require.NoError(t, err)

	//Check if new pods are coming up
	_, err = e2e.PollForTarget(ctx, e, extdeployment.DeploymentTargetType, func(target discovery_kit_api.Target) bool {
		log.Debug().Msgf("pod: %v", target.Attributes["k8s.pod.name"])
		return e2e.HasAttribute(target, "k8s.deployment", "nginx-test-delete-pod") &&
			len(target.Attributes["k8s.pod.name"]) == 2 &&
			podName1 != target.Attributes["k8s.pod.name"][0] &&
			podName2 != target.Attributes["k8s.pod.name"][0]
	})
	require.NoError(t, err)
}

func testDrainNode(t *testing.T, m *e2e.Minikube, e *e2e.Extension) {
	log.Info().Msg("Starting testDrainNode")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Start Deployment with 2 pods
	nginxDeployment := e2e.NginxDeployment{Minikube: m}
	err := nginxDeployment.Deploy("nginx-test-drain")
	require.NoError(t, err, "failed to create deployment")
	defer func() { _ = nginxDeployment.Delete() }()
	podName1 := nginxDeployment.Pods[0].Name
	podName2 := nginxDeployment.Pods[1].Name
	assert.Len(t, nginxDeployment.Pods, 2)
	log.Info().Msgf("Pods before Attack: podName1: %v, podName2: %v", podName1, podName2)

	//Check if node discovery is working and listing both pods
	nodeTarget, err := e2e.PollForTarget(ctx, e, extnode.NodeTargetType, func(target discovery_kit_api.Target) bool {
		return slices.Contains(target.Attributes["k8s.pod.name"], podName1) && slices.Contains(target.Attributes["k8s.pod.name"], podName2)
	})
	require.NoError(t, err)

	//Drain node
	config := struct {
		Duration int `json:"duration"`
	}{
		Duration: 10000,
	}
	_, err = e.RunAction(extnode.DrainNodeActionId, &action_kit_api.Target{
		Name: nodeTarget.Id,
		Attributes: map[string][]string{
			"host.hostname": nodeTarget.Attributes["host.hostname"],
		},
	}, config, nil)
	require.NoError(t, err)

	// pods are removed
	_, err = e2e.PollForTarget(ctx, e, extnode.NodeTargetType, func(target discovery_kit_api.Target) bool {
		return !slices.Contains(target.Attributes["k8s.pod.name"], podName1) && !slices.Contains(target.Attributes["k8s.pod.name"], podName2)
	})
	require.NoError(t, err)
	log.Info().Msgf("pods are removed")

	// pods are rescheduled after attack
	_, err = e2e.PollForTarget(ctx, e, extnode.NodeTargetType, func(target discovery_kit_api.Target) bool {
		for _, pod := range target.Attributes["k8s.pod.name"] {
			if strings.HasPrefix(pod, "nginx-test-drain-") {
				return true
			}
		}
		return false
	})
	log.Info().Msgf("pods are rescheduled")
	require.NoError(t, err)
}

func testTaintNode(t *testing.T, m *e2e.Minikube, e *e2e.Extension) {
	log.Info().Msg("Starting testTaintNode")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Start Deployment with 2 pods
	nginxDeployment := e2e.NginxDeployment{Minikube: m}
	err := nginxDeployment.Deploy("nginx-test-taint")
	require.NoError(t, err, "failed to create deployment")
	defer func() { _ = nginxDeployment.Delete() }()
	podName1 := nginxDeployment.Pods[0].Name
	podName2 := nginxDeployment.Pods[1].Name
	assert.Len(t, nginxDeployment.Pods, 2)
	log.Info().Msgf("Pods before Attack: podName1: %v, podName2: %v", podName1, podName2)

	//Check if node discovery is working and listing both pods
	nodeTarget, err := e2e.PollForTarget(ctx, e, extnode.NodeTargetType, func(target discovery_kit_api.Target) bool {
		return slices.Contains(target.Attributes["k8s.pod.name"], podName1) && slices.Contains(target.Attributes["k8s.pod.name"], podName2)
	})
	require.NoError(t, err)

	//Taint node
	config := struct {
		Duration int    `json:"duration"`
		Key      string `json:"key"`
		Value    string `json:"value"`
		Effect   string `json:"effect"`
	}{
		Duration: 10000,
		Key:      "allowed",
		Value:    "nothing",
		Effect:   "NoSchedule",
	}
	_, err = e.RunAction(extnode.DrainNodeActionId, &action_kit_api.Target{
		Name: nodeTarget.Id,
		Attributes: map[string][]string{
			"host.hostname": nodeTarget.Attributes["host.hostname"],
		},
	}, config, nil)
	require.NoError(t, err)
	attackStarted := time.Now()

	//Delete both pods
	_, err = e.RunAction(extpod.DeletePodActionId, &action_kit_api.Target{
		Name:       podName1,
		Attributes: map[string][]string{"k8s.pod.name": {podName1}, "k8s.namespace": {"default"}},
	}, nil, nil)
	require.NoError(t, err)
	_, err = e.RunAction(extpod.DeletePodActionId, &action_kit_api.Target{
		Name:       podName1,
		Attributes: map[string][]string{"k8s.pod.name": {podName2}, "k8s.namespace": {"default"}},
	}, nil, nil)
	require.NoError(t, err)

	// pods are removed and do not come back as long as the node is tainted
	_, err = e2e.PollForTarget(ctx, e, extnode.NodeTargetType, func(target discovery_kit_api.Target) bool {
		containsNginxPod := false
		for _, pod := range target.Attributes["k8s.pod.name"] {
			if strings.HasPrefix(pod, "nginx-test-taint-") {
				containsNginxPod = true
			}
		}
		return time.Since(attackStarted).Seconds() > 5 && !containsNginxPod
	})
	require.NoError(t, err)
	log.Info().Msgf("pods didn't come back within 5 seconds, node seems to be tainted")

	// pods are rescheduled after attack
	_, err = e2e.PollForTarget(ctx, e, extnode.NodeTargetType, func(target discovery_kit_api.Target) bool {
		for _, pod := range target.Attributes["k8s.pod.name"] {
			if strings.HasPrefix(pod, "nginx-test-taint-") {
				return true
			}
		}
		return false
	})
	log.Info().Msgf("pods are rescheduled")
	require.NoError(t, err)
}

func testScaleDeployment(t *testing.T, m *e2e.Minikube, e *e2e.Extension) {
	log.Info().Msg("Starting testScaleDeployment")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Start Deployment with 2 pods
	nginxDeployment := e2e.NginxDeployment{Minikube: m}
	err := nginxDeployment.Deploy("nginx-test-scale")
	require.NoError(t, err, "failed to create deployment")
	defer func() { _ = nginxDeployment.Delete() }()
	assert.Len(t, nginxDeployment.Pods, 2)

	//Check if node discovery is working and listing 2 pods
	nodeTarget, err := e2e.PollForTarget(ctx, e, extnode.NodeTargetType, func(target discovery_kit_api.Target) bool {
		count := 0
		for _, pod := range target.Attributes["k8s.pod.name"] {
			if strings.HasPrefix(pod, "nginx-test-scale-") {
				count++
			}
		}
		return count == 2
	})
	require.NoError(t, err)

	//Taint node
	config := struct {
		Duration     int `json:"duration"`
		ReplicaCount int `json:"replicaCount"`
	}{
		Duration:     10000,
		ReplicaCount: 3,
	}
	_, err = e.RunAction(extdeployment.ScaleDeploymentActionId, &action_kit_api.Target{
		Name: nodeTarget.Id,
		Attributes: map[string][]string{
			"k8s.namespace":  {"default"},
			"k8s.deployment": {"nginx-test-scale"},
		},
	}, config, nil)
	require.NoError(t, err)

	// pods are upscaled
	_, err = e2e.PollForTarget(ctx, e, extnode.NodeTargetType, func(target discovery_kit_api.Target) bool {
		count := 0
		for _, pod := range target.Attributes["k8s.pod.name"] {
			if strings.HasPrefix(pod, "nginx-test-scale-") {
				count++
			}
		}
		return count == 3
	})
	require.NoError(t, err)
	log.Info().Msgf("pods are scaled to 3")

	// pod scale is reverted to 2 after attack
	_, err = e2e.PollForTarget(ctx, e, extnode.NodeTargetType, func(target discovery_kit_api.Target) bool {
		count := 0
		for _, pod := range target.Attributes["k8s.pod.name"] {
			if strings.HasPrefix(pod, "nginx-test-scale-") {
				count++
			}
		}
		return count == 2
	})
	require.NoError(t, err)
	log.Info().Msgf("pod replica count is back to 2")
}
