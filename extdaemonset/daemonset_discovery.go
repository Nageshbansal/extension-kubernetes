// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package extdaemonset

import (
	"fmt"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_commons"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/extension-kubernetes/client"
	"github.com/steadybit/extension-kubernetes/extconfig"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/utils/strings/slices"
	"net/http"
	"strings"
)

func RegisterStatefulSetDiscoveryHandlers() {
	exthttp.RegisterHttpHandler("/daemonset/discovery", exthttp.GetterAsHandler(getDaemonSetDiscoveryDescription))
	exthttp.RegisterHttpHandler("/daemonset/discovery/target-description", exthttp.GetterAsHandler(getDaemonSetTargetDescription))
	exthttp.RegisterHttpHandler("/daemonset/discovery/discovered-targets", getDiscoveredDaemonSets)
	exthttp.RegisterHttpHandler("/daemonset/discovery/rules/k8s-daemonset-to-container", exthttp.GetterAsHandler(getDaemonSetToContainerEnrichmentRule))
}

func getDaemonSetDiscoveryDescription() discovery_kit_api.DiscoveryDescription {
	return discovery_kit_api.DiscoveryDescription{
		Id:         DaemonSetTargetType,
		RestrictTo: extutil.Ptr(discovery_kit_api.LEADER),
		Discover: discovery_kit_api.DescribingEndpointReferenceWithCallInterval{
			Method:       "GET",
			Path:         "/daemonset/discovery/discovered-targets",
			CallInterval: extutil.Ptr("1m"),
		},
	}
}

func getDaemonSetTargetDescription() discovery_kit_api.TargetDescription {
	return discovery_kit_api.TargetDescription{
		Id:       DaemonSetTargetType,
		Label:    discovery_kit_api.PluralLabel{One: "Kubernetes DaemonSet", Other: "Kubernetes DaemonSets"},
		Category: extutil.Ptr("Kubernetes"),
		Version:  extbuild.GetSemverVersionStringOrUnknown(),
		Icon:     extutil.Ptr(daemonSetIcon),
		Table: discovery_kit_api.Table{
			Columns: []discovery_kit_api.Column{
				{Attribute: "k8s.daemonset"},
				{Attribute: "k8s.namespace"},
				{Attribute: "k8s.cluster-name"},
			},
			OrderBy: []discovery_kit_api.OrderBy{
				{
					Attribute: "k8s.daemonset",
					Direction: "ASC",
				},
			},
		},
	}
}

func getDiscoveredDaemonSets(w http.ResponseWriter, _ *http.Request, _ []byte) {
	targets := getDiscoveredDaemonSetTargets(client.K8S)
	exthttp.WriteBody(w, discovery_kit_api.DiscoveryData{Targets: &targets})
}

func getDiscoveredDaemonSetTargets(k8s *client.Client) []discovery_kit_api.Target {
	daemonsets := k8s.DaemonSets()

	filteredDaemonSets := make([]*appsv1.DaemonSet, 0, len(daemonsets))
	if extconfig.Config.DisableDiscoveryExcludes {
		filteredDaemonSets = daemonsets
	} else {
		for _, ds := range daemonsets {
			if client.IsExcludedFromDiscovery(ds.ObjectMeta) {
				continue
			}
			filteredDaemonSets = append(filteredDaemonSets, ds)
		}
	}

	targets := make([]discovery_kit_api.Target, len(filteredDaemonSets))
	for i, ds := range filteredDaemonSets {
		targetName := fmt.Sprintf("%s/%s/%s", extconfig.Config.ClusterName, ds.Namespace, ds.Name)
		attributes := map[string][]string{
			"k8s.namespace":    {ds.Namespace},
			"k8s.daemonset":    {ds.Name},
			"k8s.cluster-name": {extconfig.Config.ClusterName},
			"k8s.distribution": {k8s.Distribution},
		}

		for key, value := range ds.ObjectMeta.Labels {
			if !slices.Contains(extconfig.Config.LabelFilter, key) {
				attributes[fmt.Sprintf("k8s.label.%v", key)] = []string{value}
			}
		}

		pods := k8s.PodsByLabelSelector(ds.Spec.Selector, ds.Namespace)
		if len(pods) > 0 {
			podNames := make([]string, len(pods))
			var containerIds []string
			var containerIdsWithoutPrefix []string
			var hostnames []string
			for podIndex, pod := range pods {
				podNames[podIndex] = pod.Name
				for _, container := range pod.Status.ContainerStatuses {
					if container.ContainerID == "" {
						continue
					}
					containerIds = append(containerIds, container.ContainerID)
					containerIdsWithoutPrefix = append(containerIdsWithoutPrefix, strings.SplitAfter(container.ContainerID, "://")[1])
				}
				hostnames = append(hostnames, pod.Spec.NodeName)
			}
			attributes["k8s.pod.name"] = podNames
			if len(containerIds) > 0 {
				attributes["k8s.container.id"] = containerIds
			}
			if len(containerIdsWithoutPrefix) > 0 {
				attributes["k8s.container.id.stripped"] = containerIdsWithoutPrefix
			}
			if len(hostnames) > 0 {
				attributes["host.hostname"] = hostnames
			}
		}

		targets[i] = discovery_kit_api.Target{
			Id:         targetName,
			TargetType: DaemonSetTargetType,
			Label:      ds.Name,
			Attributes: attributes,
		}
	}
	return discovery_kit_commons.ApplyAttributeExcludes(targets, extconfig.Config.DiscoveryAttributesExcludesDaemonSet)
}

func getDaemonSetToContainerEnrichmentRule() discovery_kit_api.TargetEnrichmentRule {
	return discovery_kit_api.TargetEnrichmentRule{
		Id:      "com.steadybit.extension_kubernetes.kubernetes-daemonset-to-container",
		Version: extbuild.GetSemverVersionStringOrUnknown(),
		Src: discovery_kit_api.SourceOrDestination{
			Type: DaemonSetTargetType,
			Selector: map[string]string{
				"k8s.container.id.stripped": "${dest.container.id.stripped}",
			},
		},
		Dest: discovery_kit_api.SourceOrDestination{
			Type: "com.steadybit.extension_container.container",
			Selector: map[string]string{
				"container.id.stripped": "${src.k8s.container.id.stripped}",
			},
		},
		Attributes: []discovery_kit_api.Attribute{
			{
				Matcher: discovery_kit_api.StartsWith,
				Name:    "k8s.label.",
			},
		},
	}
}