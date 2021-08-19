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

	"github.com/openkruise/kruise-api/apps/v1beta1"

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
	descStatefulSetLabelsName          = "kruise_statefulset_labels"
	descStatefulSetLabelsHelp          = "Kruise labels converted to Prometheus labels."
	descStatefulSetLabelsDefaultLabels = []string{"namespace", "statefulset"}
)

func statefulSetMetricFamilies(allowLabelsList []string) []generator.FamilyGenerator {
	return []generator.FamilyGenerator{
		*generator.NewFamilyGenerator(
			"kruise_statefulset_created",
			"Unix creation timestamp",
			metric.Gauge,
			"",
			wrapStatefulSetFunc(func(s *v1beta1.StatefulSet) *metric.Family {
				ms := []*metric.Metric{}

				if !s.CreationTimestamp.IsZero() {
					ms = append(ms, &metric.Metric{
						Value: float64(s.CreationTimestamp.Unix()),
					})
				}

				return &metric.Family{
					Metrics: ms,
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_statefulset_status_replicas",
			"The number of replicas per statefulset",
			metric.Gauge,
			"",
			wrapStatefulSetFunc(func(s *v1beta1.StatefulSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(s.Status.Replicas),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_statefulset_status_replicas_available",
			"The number of available replicas per statefulset",
			metric.Gauge,
			"",
			wrapStatefulSetFunc(func(s *v1beta1.StatefulSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(s.Status.AvailableReplicas),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_statefulset_status_replicas_current",
			"The number of current replicas per statefulset",
			metric.Gauge,
			"",
			wrapStatefulSetFunc(func(s *v1beta1.StatefulSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(s.Status.CurrentReplicas),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_statefulset_status_replicas_ready",
			"The number of ready replicas per statefulset",
			metric.Gauge,
			"",
			wrapStatefulSetFunc(func(s *v1beta1.StatefulSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(s.Status.ReadyReplicas),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_statefulset_status_replicas_updated",
			"The number of updated replicas per statefulset",
			metric.Gauge,
			"",
			wrapStatefulSetFunc(func(s *v1beta1.StatefulSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(s.Status.UpdatedReplicas),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_statefulset_status_observed_generation",
			"The generation observed by the statefulset controller.",
			metric.Gauge,
			"",
			wrapStatefulSetFunc(func(s *v1beta1.StatefulSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(s.Status.ObservedGeneration),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_statefulset_status_condition",
			"The current status conditions of a statefulset.",
			metric.Gauge,
			"",
			wrapStatefulSetFunc(func(cs *v1beta1.StatefulSet) *metric.Family {
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
			"kruise_statefulset_replicas",
			"Number of desired pods for a statefulset.",
			metric.Gauge,
			"",
			wrapStatefulSetFunc(func(s *v1beta1.StatefulSet) *metric.Family {
				ms := []*metric.Metric{}

				if s.Spec.Replicas != nil {
					ms = append(ms, &metric.Metric{
						Value: float64(*s.Spec.Replicas),
					})
				}

				return &metric.Family{
					Metrics: ms,
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_statefulset_metadata_generation",
			"Sequence number representing a specific generation of the desired state for the statefulset.",
			metric.Gauge,
			"",
			wrapStatefulSetFunc(func(s *v1beta1.StatefulSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(s.ObjectMeta.Generation),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_statefulset_spec_replicas",
			"Number of desired pods for a statefulset.",
			metric.Gauge,
			"",
			wrapStatefulSetFunc(func(s *v1beta1.StatefulSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(*s.Spec.Replicas),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_statefulset_spec_strategy_rollingupdate_max_unavailable",
			"Maximum number of unavailable replicas during a rolling update of a statefulset.",
			metric.Gauge,
			"",
			wrapStatefulSetFunc(func(cs *v1beta1.StatefulSet) *metric.Family {
				maxUnavailable, err := intstr.GetValueFromIntOrPercent(cs.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable, int(*cs.Spec.Replicas), false)
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
			"kruise_statefulset_spec_reserveordinals",
			"Maximum number of replicas that can be scheduled above the desired number of replicas during a rolling update of a statefulset.",
			metric.Gauge,
			"",
			wrapStatefulSetFunc(func(cs *v1beta1.StatefulSet) *metric.Family {
				ms := make([]*metric.Metric, len(cs.Spec.ReserveOrdinals))
				for i, m := range cs.Spec.ReserveOrdinals {
					ms[i].Value = float64(m)
				}

				return &metric.Family{
					Metrics: ms,
				}
			}),
		),
		*generator.NewFamilyGenerator(
			descStatefulSetLabelsName,
			descStatefulSetLabelsHelp,
			metric.Gauge,
			"",
			wrapStatefulSetFunc(func(cs *v1beta1.StatefulSet) *metric.Family {
				labelKeys, labelValues := createLabelKeysValues(cs.Labels, allowLabelsList)
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
			"kruise_statefulset_status_current_revision",
			"Indicates the version of the statefulset used to generate Pods in the sequence [0,currentReplicas).",
			metric.Gauge,
			"",
			wrapStatefulSetFunc(func(s *v1beta1.StatefulSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							LabelKeys:   []string{"revision"},
							LabelValues: []string{s.Status.CurrentRevision},
							Value:       1,
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_statefulset_status_update_revision",
			"Indicates the version of the statefulset used to generate Pods in the sequence [replicas-updatedReplicas,replicas)",
			metric.Gauge,
			"",
			wrapStatefulSetFunc(func(s *v1beta1.StatefulSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							LabelKeys:   []string{"revision"},
							LabelValues: []string{s.Status.UpdateRevision},
							Value:       1,
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_statefulset_spec_strategy_type",
			"The type of updateStrategy.",
			metric.Gauge,
			"",
			wrapStatefulSetFunc(func(s *v1beta1.StatefulSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							LabelKeys:   []string{"strategy_type"},
							LabelValues: []string{string(s.Spec.UpdateStrategy.Type)},
						},
					},
				}
			}),
		),
	}
}
func wrapStatefulSetFunc(f func(*v1beta1.StatefulSet) *metric.Family) func(interface{}) *metric.Family {
	return func(obj interface{}) *metric.Family {
		statefulset := obj.(*v1beta1.StatefulSet)

		metricFamily := f(statefulset)

		for _, m := range metricFamily.Metrics {
			m.LabelKeys = append(descStatefulSetLabelsDefaultLabels, m.LabelKeys...)
			m.LabelValues = append([]string{statefulset.Namespace, statefulset.Name}, m.LabelValues...)
		}

		return metricFamily
	}
}

func createStatefulSetListWatch(kruiseClient kruiseclientset.Interface, ns string) cache.ListerWatcher {
	return &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return kruiseClient.AppsV1alpha1().StatefulSets(ns).List(context.TODO(), opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			return kruiseClient.AppsV1alpha1().StatefulSets(ns).Watch(context.TODO(), opts)
		},
	}
}
