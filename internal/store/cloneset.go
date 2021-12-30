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
	descCloneSetAnnotationsName     = "kruise_cloneset_annotations"
	descCloneSetAnnotationsHelp     = "Kruise annotations converted to Prometheus labels."
	descCloneSetLabelsName          = "kruise_cloneset_labels"
	descCloneSetLabelsHelp          = "Kruise labels converted to Prometheus labels."
	descCloneSetLabelsDefaultLabels = []string{"namespace", "cloneset"}

	cloneSetUpdateTypes = []v1alpha1.CloneSetUpdateStrategyType{
		v1alpha1.RecreateCloneSetUpdateStrategyType,
		v1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType,
		v1alpha1.InPlaceOnlyCloneSetUpdateStrategyType,
	}
)

func cloneSetMetricFamilies(allowAnnotationsList, allowLabelsList []string) []generator.FamilyGenerator {
	return []generator.FamilyGenerator{
		*generator.NewFamilyGenerator(
			"kruise_cloneset_created",
			"Unix creation timestamp",
			metric.Gauge,
			"",
			wrapCloneSetFunc(func(cs *v1alpha1.CloneSet) *metric.Family {
				ms := []*metric.Metric{}

				if !cs.CreationTimestamp.IsZero() {
					ms = append(ms, &metric.Metric{
						Value: float64(cs.CreationTimestamp.Unix()),
					})
				}

				return &metric.Family{
					Metrics: ms,
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_cloneset_status_replicas",
			"The number of replicas per cloneset.",
			metric.Gauge,
			"",
			wrapCloneSetFunc(func(cs *v1alpha1.CloneSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(cs.Status.Replicas),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_cloneset_status_replicas_available",
			"The number of available replicas per cloneset.",
			metric.Gauge,
			"",
			wrapCloneSetFunc(func(cs *v1alpha1.CloneSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(cs.Status.AvailableReplicas),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_cloneset_status_replicas_updated",
			"The number of updated replicas per cloneset.",
			metric.Gauge,
			"",
			wrapCloneSetFunc(func(cs *v1alpha1.CloneSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(cs.Status.UpdatedReplicas),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_cloneset_status_observed_generation",
			"The generation observed by the cloneset controller.",
			metric.Gauge,
			"",
			wrapCloneSetFunc(func(cs *v1alpha1.CloneSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(cs.Status.ObservedGeneration),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_cloneset_status_condition",
			"The current status conditions of a cloneset.",
			metric.Gauge,
			"",
			wrapCloneSetFunc(func(cs *v1alpha1.CloneSet) *metric.Family {
				ms := make([]*metric.Metric, len(cs.Status.Conditions)*len(conditionStatuses))

				for i, cs := range cs.Status.Conditions {
					conditionMetrics := addConditionMetrics(cs.Status)

					for j, m := range conditionMetrics {
						metric := m

						metric.LabelKeys = []string{"condition", "status"}
						metric.LabelValues = append([]string{string(cs.Type)}, metric.LabelValues...)
						ms[i*len(conditionStatuses)+j] = metric
					}
				}

				return &metric.Family{
					Metrics: ms,
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_cloneset_spec_replicas",
			"Number of desired pods for a cloneset.",
			metric.Gauge,
			"",
			wrapCloneSetFunc(func(cs *v1alpha1.CloneSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(*cs.Spec.Replicas),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_cloneset_spec_strategy_rollingupdate_max_unavailable",
			"Maximum number of unavailable replicas during a rolling update of a cloneset.",
			metric.Gauge,
			"",
			wrapCloneSetFunc(func(cs *v1alpha1.CloneSet) *metric.Family {
				updateStrategyMaxUnavailable := intstr.FromInt(0)
				if cs.Spec.UpdateStrategy.MaxUnavailable == nil {
					updateStrategyMaxUnavailable = *cs.Spec.UpdateStrategy.MaxUnavailable
				}
				maxUnavailable, err := intstr.GetValueFromIntOrPercent(&updateStrategyMaxUnavailable, int(*cs.Spec.Replicas), false)
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
			"kruise_cloneset_spec_strategy_rollingupdate_max_surge",
			"Maximum number of replicas that can be scheduled above the desired number of replicas during a rolling update of a cloneset.",
			metric.Gauge,
			"",
			wrapCloneSetFunc(func(cs *v1alpha1.CloneSet) *metric.Family {
				updateStrategyMaxSurge := intstr.FromInt(0)
				if cs.Spec.UpdateStrategy.MaxSurge != nil {
					updateStrategyMaxSurge = *cs.Spec.UpdateStrategy.MaxSurge
				}
				maxSurge, err := intstr.GetValueFromIntOrPercent(&updateStrategyMaxSurge, int(*cs.Spec.Replicas), true)
				if err != nil {
					panic(err)
				}

				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(maxSurge),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_cloneset_metadata_generation",
			"Sequence number representing a specific generation of the desired state.",
			metric.Gauge,
			"",
			wrapCloneSetFunc(func(cs *v1alpha1.CloneSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(cs.ObjectMeta.Generation),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			descCloneSetAnnotationsName,
			descCloneSetAnnotationsHelp,
			metric.Gauge,
			"",
			wrapCloneSetFunc(func(cs *v1alpha1.CloneSet) *metric.Family {
				annotationKeys, annotationValues := createPrometheusLabelKeysValues("annotation", cs.Annotations, allowAnnotationsList)
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
			descCloneSetLabelsName,
			descCloneSetLabelsHelp,
			metric.Gauge,
			"",
			wrapCloneSetFunc(func(cs *v1alpha1.CloneSet) *metric.Family {
				labelKeys, labelValues := createPrometheusLabelKeysValues("label", cs.Labels, allowLabelsList)
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
			"kruise_cloneset_status_replicas_ready",
			"The number of ready replicas per cloneset.",
			metric.Gauge,
			"",
			wrapCloneSetFunc(func(cs *v1alpha1.CloneSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(cs.Status.ReadyReplicas),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_cloneset_status_replicas_updated_ready",
			"The number of update and ready replicas per cloneset.",
			metric.Gauge,
			"",
			wrapCloneSetFunc(func(cs *v1alpha1.CloneSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(cs.Status.UpdatedReadyReplicas),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_cloneset_spec_strategy_partition",
			"Desired number or percent of Pods in old revisions.",
			metric.Gauge,
			"",
			wrapCloneSetFunc(func(cs *v1alpha1.CloneSet) *metric.Family {
				if cs.Spec.UpdateStrategy.Partition == nil {
					return &metric.Family{}
				}

				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							LabelKeys:   []string{"partition"},
							LabelValues: []string{cs.Spec.UpdateStrategy.Partition.String()},
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_cloneset_spec_strategy_type",
			"The type of updateStrategy.",
			metric.Gauge,
			"",
			wrapCloneSetFunc(func(cs *v1alpha1.CloneSet) *metric.Family {
				ms := make([]*metric.Metric, len(cloneSetUpdateTypes))

				for i, t := range cloneSetUpdateTypes {
					ms[i] = &metric.Metric{
						LabelKeys:   []string{"strategy_type"},
						LabelValues: []string{string(t)},
						Value:       boolFloat64(cs.Spec.UpdateStrategy.Type == t),
					}
				}

				return &metric.Family{
					Metrics: ms,
				}
			}),
		),
	}
}

func wrapCloneSetFunc(f func(*v1alpha1.CloneSet) *metric.Family) func(interface{}) *metric.Family {
	return func(obj interface{}) *metric.Family {
		cloneset := obj.(*v1alpha1.CloneSet)

		metricFamily := f(cloneset)

		for _, m := range metricFamily.Metrics {
			m.LabelKeys = append(descCloneSetLabelsDefaultLabels, m.LabelKeys...)
			m.LabelValues = append([]string{cloneset.Namespace, cloneset.Name}, m.LabelValues...)
		}

		return metricFamily
	}
}

func createCloneSetListWatch(kruiseClient kruiseclientset.Interface, ns string) cache.ListerWatcher {
	return &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return kruiseClient.AppsV1alpha1().CloneSets(ns).List(context.TODO(), opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			return kruiseClient.AppsV1alpha1().CloneSets(ns).Watch(context.TODO(), opts)
		},
	}
}
