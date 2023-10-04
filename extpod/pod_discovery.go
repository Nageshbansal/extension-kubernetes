// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extpod

import (
	"fmt"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_commons"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/extension-kubernetes/client"
	"github.com/steadybit/extension-kubernetes/extconfig"
	"golang.org/x/exp/slices"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"strings"
)

func RegisterPodDiscoveryHandlers() {
	exthttp.RegisterHttpHandler("/pod/discovery", exthttp.GetterAsHandler(getPodDiscoveryDescription))
	exthttp.RegisterHttpHandler("/pod/discovery/target-description", exthttp.GetterAsHandler(getPodTargetDescription))
	exthttp.RegisterHttpHandler("/pod/discovery/discovered-targets", getDiscoveredPods)
}

func getPodDiscoveryDescription() discovery_kit_api.DiscoveryDescription {
	return discovery_kit_api.DiscoveryDescription{
		Id:         PodTargetType,
		RestrictTo: extutil.Ptr(discovery_kit_api.LEADER),
		Discover: discovery_kit_api.DescribingEndpointReferenceWithCallInterval{
			Method:       "GET",
			Path:         "/pod/discovery/discovered-targets",
			CallInterval: extutil.Ptr("1m"),
		},
	}
}

func getPodTargetDescription() discovery_kit_api.TargetDescription {
	return discovery_kit_api.TargetDescription{
		Id:       PodTargetType,
		Label:    discovery_kit_api.PluralLabel{One: "Kubernetes Pod", Other: "Kubernetes Pods"},
		Category: extutil.Ptr("Kubernetes"),
		Version:  extbuild.GetSemverVersionStringOrUnknown(),
		Icon:     extutil.Ptr(podIcon),
		Table: discovery_kit_api.Table{
			Columns: []discovery_kit_api.Column{
				{Attribute: "k8s.pod.name"},
				{Attribute: "k8s.cluster-name"},
				{Attribute: "k8s.namespace"},
				{Attribute: "k8s.deployment", FallbackAttributes: extutil.Ptr([]string{"k8s.statefulset", "k8s.daemonset"})},
			},
			OrderBy: []discovery_kit_api.OrderBy{
				{
					Attribute: "k8s.cluster-name",
					Direction: "ASC",
				},
			},
		},
	}
}

func getDiscoveredPods(w http.ResponseWriter, _ *http.Request, _ []byte) {
	targets := getDiscoveredPodTargets(client.K8S)
	exthttp.WriteBody(w, discovery_kit_api.DiscoveryData{Targets: &targets})
}
func getDiscoveredPodTargets(k8s *client.Client) []discovery_kit_api.Target {
	pods := k8s.Pods()

	filteredPods := make([]*corev1.Pod, 0, len(pods))
	if extconfig.Config.DisableDiscoveryExcludes {
		filteredPods = pods
	} else {
		for _, d := range pods {
			if client.IsExcludedFromDiscovery(d.ObjectMeta) {
				continue
			}
			filteredPods = append(filteredPods, d)
		}
	}

	targets := make([]discovery_kit_api.Target, len(filteredPods))
	for i, p := range filteredPods {
		attributes := map[string][]string{
			"k8s.pod.name":     {p.Name},
			"k8s.namespace":    {p.Namespace},
			"k8s.cluster-name": {extconfig.Config.ClusterName},
			"k8s.node.name":    {p.Spec.NodeName},
		}

		for key, value := range p.ObjectMeta.Labels {
			if !slices.Contains(extconfig.Config.LabelFilter, key) {
				attributes[fmt.Sprintf("k8s.label.%v", key)] = []string{value}
			}
		}

		var containerIds []string
		var containerIdsWithoutPrefix []string
		for _, container := range p.Status.ContainerStatuses {
			if container.ContainerID == "" {
				continue
			}
			containerIds = append(containerIds, container.ContainerID)
			containerIdsWithoutPrefix = append(containerIdsWithoutPrefix, strings.SplitAfter(container.ContainerID, "://")[1])
		}
		if len(containerIds) > 0 {
			attributes["k8s.container.id"] = containerIds
		}
		if len(containerIdsWithoutPrefix) > 0 {
			attributes["k8s.container.id.stripped"] = containerIdsWithoutPrefix
		}

		ownerReferences := client.OwnerReferences(k8s, &p.ObjectMeta)
		for _, ownerRef := range ownerReferences.OwnerRefs {
			attributes[fmt.Sprintf("k8s.%v", ownerRef.Kind)] = []string{ownerRef.Name}
		}

		targets[i] = discovery_kit_api.Target{
			Id:         p.Name,
			TargetType: PodTargetType,
			Label:      p.Name,
			Attributes: attributes,
		}
	}
	return discovery_kit_commons.ApplyAttributeExcludes(targets, extconfig.Config.DiscoveryAttributesExcludesPod)
}
