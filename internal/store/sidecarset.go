/*
Copyright 2021 The Kruise Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package store

import (
	"context"

	"github.com/openkruise/kruise-api/apps/v1alpha1"

	kruiseclientset "github.com/openkruise/kruise-api/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"k8s.io/kube-state-metrics/v2/pkg/metric"
	generator "k8s.io/kube-state-metrics/v2/pkg/metric_generator"
)

var (
	descSidecarSetAnnotationsName     = "kruise_sidecarset_annotations"
	descSidecarSetAnnotationsHelp     = "Kruise annotations converted to Prometheus labels."
	descSidecarSetLabelsName          = "kruise_sidecarset_labels"
	descSidecarSetLabelsHelp          = "Kruise labels converted to Prometheus labels."
	descSidecarSetLabelsDefaultLabels = []string{"namespace", "sidecarset"}
)

func sidecarSetMetricFamilies(allowAnnotationsList, allowLabelsList []string) []generator.FamilyGenerator {
	return []generator.FamilyGenerator{
		*generator.NewFamilyGenerator(
			"kruise_sidecarset_created",
			"Unix creation timestamp",
			metric.Gauge,
			"",
			wrapSidecarSetFunc(func(sc *v1alpha1.SidecarSet) *metric.Family {
				ms := []*metric.Metric{}

				if !sc.CreationTimestamp.IsZero() {
					ms = append(ms, &metric.Metric{
						Value: float64(sc.CreationTimestamp.Unix()),
					})
				}

				return &metric.Family{
					Metrics: ms,
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_sidecarset_status_replicas_matched",
			"The number of matched replicas per sidecarset.",
			metric.Gauge,
			"",
			wrapSidecarSetFunc(func(sc *v1alpha1.SidecarSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(sc.Status.MatchedPods),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"ruise_sidecarset_status_replicas_updated",
			"The number of updated replicas per sidecarset.",
			metric.Gauge,
			"",
			wrapSidecarSetFunc(func(sc *v1alpha1.SidecarSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(sc.Status.UpdatedPods),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_sidecarset_status_replicas_ready ",
			"The number of ready  replicas per sidecarset.",
			metric.Gauge,
			"",
			wrapSidecarSetFunc(func(sc *v1alpha1.SidecarSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(sc.Status.ReadyPods),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_sidecarset_status_observed_generation",
			"The generation observed by the sidecarset controller.",
			metric.Gauge,
			"",
			wrapSidecarSetFunc(func(sc *v1alpha1.SidecarSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(sc.Status.ObservedGeneration),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_sidecarset_status_replicas_updated_ready",
			"The number of update and ready replicas per sidecarset.",
			metric.Gauge,
			"",
			wrapSidecarSetFunc(func(sc *v1alpha1.SidecarSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(sc.Status.UpdatedReadyPods),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_sidecarset_spec_namespcace",
			"The namespace matched pods in.",
			metric.Gauge,
			"",
			wrapSidecarSetFunc(func(sc *v1alpha1.SidecarSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							LabelKeys:   []string{"namespace"},
							LabelValues: []string{sc.Spec.Namespace},
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_sidecarset_spec_strategy_rollingupdate_max_unavailable",
			"Maximum number of unavailable replicas during a rolling update of a sidecarset.",
			metric.Gauge,
			"",
			wrapSidecarSetFunc(func(sc *v1alpha1.SidecarSet) *metric.Family {
				maxUnavailable, err := intstr.GetValueFromIntOrPercent(sc.Spec.UpdateStrategy.MaxUnavailable, int(sc.Status.MatchedPods), false)
				if err != nil {
					panic(err)
				}

				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(maxUnavailable),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_sidecarset_spec_strategy_partition",
			"Desired number or percent of Pods in old revisions.",
			metric.Gauge,
			"",
			wrapSidecarSetFunc(func(sc *v1alpha1.SidecarSet) *metric.Family {
				if sc.Spec.UpdateStrategy.Partition == nil {
					return &metric.Family{}
				}

				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							LabelKeys:   []string{"partition"},
							LabelValues: []string{sc.Spec.UpdateStrategy.Partition.String()},
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_sidecarset_spec_strategy_type",
			"The type of updateStrategy.",
			metric.Gauge,
			"",
			wrapSidecarSetFunc(func(sc *v1alpha1.SidecarSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							LabelKeys:   []string{"strategy_type"},
							LabelValues: []string{string(sc.Spec.UpdateStrategy.Type)},
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_sidecarset_spec_metadata_generation",
			"Sequence number representing a specific generation of the desired state.",
			metric.Gauge,
			"",
			wrapSidecarSetFunc(func(sc *v1alpha1.SidecarSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(sc.ObjectMeta.Generation),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			descSidecarSetAnnotationsName,
			descSidecarSetAnnotationsHelp,
			metric.Gauge,
			"",
			wrapSidecarSetFunc(func(sc *v1alpha1.SidecarSet) *metric.Family {
				annotationKeys, annotationValues := createPrometheusLabelKeysValues("annotation", sc.Annotations, allowAnnotationsList)
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							LabelKeys:   annotationKeys,
							LabelValues: annotationValues,
							Value:       1,
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			descSidecarSetLabelsName,
			descSidecarSetLabelsHelp,
			metric.Gauge,
			"",
			wrapSidecarSetFunc(func(sc *v1alpha1.SidecarSet) *metric.Family {
				labelKeys, labelValues := createPrometheusLabelKeysValues("label", sc.Labels, allowLabelsList)
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							LabelKeys:   labelKeys,
							LabelValues: labelValues,
							Value:       1,
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_sidecarset_spec_containers_injectpolicy",
			"The rules that injected SidecarContainer into Pod.spec.containers.",
			metric.Gauge,
			"",
			wrapSidecarSetFunc(func(sc *v1alpha1.SidecarSet) *metric.Family {
				ms := []*metric.Metric{}
				for _, container := range sc.Spec.Containers {
					ms = append(ms, &metric.Metric{
						LabelKeys:   []string{"injectpolicy"},
						LabelValues: []string{string(container.PodInjectPolicy)},
					})
				}
				return &metric.Family{
					Metrics: ms,
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_sidecarset_spec_containers_strategy_type",
			"The type of containers' upgradeStrategy.",
			metric.Gauge,
			"",
			wrapSidecarSetFunc(func(sc *v1alpha1.SidecarSet) *metric.Family {
				ms := []*metric.Metric{}
				for _, container := range sc.Spec.Containers {
					ms = append(ms, &metric.Metric{
						LabelKeys:   []string{"strategy_type"},
						LabelValues: []string{string(container.UpgradeStrategy.UpgradeType)},
					})
				}
				return &metric.Family{
					Metrics: ms,
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_sidecarset_spec_containers_strategy_hotupgradeemptyimage",
			"The consistent of sidecar container.",
			metric.Gauge,
			"",
			wrapSidecarSetFunc(func(sc *v1alpha1.SidecarSet) *metric.Family {
				ms := []*metric.Metric{}
				for _, container := range sc.Spec.Containers {
					ms = append(ms, &metric.Metric{
						LabelKeys:   []string{"hotupgradeemptyimage"},
						LabelValues: []string{container.UpgradeStrategy.HotUpgradeEmptyImage},
					})
				}
				return &metric.Family{
					Metrics: ms,
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_sidecarset_spec_containers_volumepolicy",
			"The other container's VolumeMounts shared.",
			metric.Gauge,
			"",
			wrapSidecarSetFunc(func(sc *v1alpha1.SidecarSet) *metric.Family {
				ms := []*metric.Metric{}
				for _, container := range sc.Spec.Containers {
					ms = append(ms, &metric.Metric{
						LabelKeys:   []string{"volumepolicy"},
						LabelValues: []string{string(container.ShareVolumePolicy.Type)},
					})
				}
				return &metric.Family{
					Metrics: ms,
				}
			}),
		),
	}
}

func wrapSidecarSetFunc(f func(*v1alpha1.SidecarSet) *metric.Family) func(interface{}) *metric.Family {
	return func(obj interface{}) *metric.Family {
		sidecarset := obj.(*v1alpha1.SidecarSet)

		metricFamily := f(sidecarset)

		for _, m := range metricFamily.Metrics {
			m.LabelKeys = append(descSidecarSetLabelsDefaultLabels, m.LabelKeys...)
			m.LabelValues = append([]string{sidecarset.Namespace, sidecarset.Name}, m.LabelValues...)
		}

		return metricFamily
	}
}

func createSidecarSetListWatch(kruiseClient kruiseclientset.Interface, ns string) cache.ListerWatcher {
	// namespace(ns) unused
	return &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return kruiseClient.AppsV1alpha1().SidecarSets().List(context.TODO(), opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			return kruiseClient.AppsV1alpha1().SidecarSets().Watch(context.TODO(), opts)
		},
	}
}
