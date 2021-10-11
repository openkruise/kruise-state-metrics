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
	"strconv"

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
	descDaemonSetAnnotationsName     = "kruise_daemonset_annotations"
	descDaemonSetAnnotationsHelp     = "Kruise annotations converted to Prometheus labels."
	descDaemonSetLabelsName          = "kruise_daemonset_labels"
	descDaemonSetLabelsHelp          = "Kruise labels converted to Prometheus labels."
	descDaemonSetLabelsDefaultLabels = []string{"namespace", "daemonset"}
)

func daemonSetMetricFamilies(allowAnnotationsList, allowLabelsList []string) []generator.FamilyGenerator {
	return []generator.FamilyGenerator{
		*generator.NewFamilyGenerator(
			"kruise_daemonset_created",
			"Unix creation timestamp",
			metric.Gauge,
			"",
			wrapDaemonSetFunc(func(ds *v1alpha1.DaemonSet) *metric.Family {
				ms := []*metric.Metric{}

				if !ds.CreationTimestamp.IsZero() {
					ms = append(ms, &metric.Metric{
						Value: float64(ds.CreationTimestamp.Unix()),
					})
				}

				return &metric.Family{
					Metrics: ms,
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_daemonset_status_condition",
			"The current status conditions of a daemonset.",
			metric.Gauge,
			"",
			wrapDaemonSetFunc(func(ds *v1alpha1.DaemonSet) *metric.Family {
				ms := make([]*metric.Metric, len(ds.Status.Conditions)*len(conditionStatuses))

				for i, ds := range ds.Status.Conditions {
					conditionMetrics := addConditionMetrics(ds.Status)

					for j, m := range conditionMetrics {
						metric := m

						metric.LabelKeys = []string{"condition", "status"}
						metric.LabelValues = append([]string{string(ds.Type)}, metric.LabelValues...)
						ms[i*len(conditionStatuses)+j] = metric
					}
				}

				return &metric.Family{
					Metrics: ms,
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_daemonset_spec_strategy_rollingupdate_max_surge",
			"Maximum number of replicas that can be scheduled above the desired number of replicas during a rolling update of a daemonset.",
			metric.Gauge,
			"",
			wrapDaemonSetFunc(func(ds *v1alpha1.DaemonSet) *metric.Family {
				maxSurge, err := intstr.GetValueFromIntOrPercent(ds.Spec.UpdateStrategy.RollingUpdate.MaxSurge, int(ds.Status.DesiredNumberScheduled), true)
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
			"kruise_daemonset_spec_strategy_partition",
			"Desired number or percent of Pods in old revisions.",
			metric.Gauge,
			"",
			wrapDaemonSetFunc(func(ds *v1alpha1.DaemonSet) *metric.Family {
				if ds.Spec.UpdateStrategy.RollingUpdate.Partition == nil {
					return &metric.Family{}
				}

				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							LabelKeys:   []string{"partition"},
							LabelValues: []string{strconv.Itoa(int(*ds.Spec.UpdateStrategy.RollingUpdate.Partition))},
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_daemonset_spec_strategy_type",
			"The type of updateStrategy.",
			metric.Gauge,
			"",
			wrapDaemonSetFunc(func(ds *v1alpha1.DaemonSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							LabelKeys:   []string{"strategy_type"},
							LabelValues: []string{string(ds.Spec.UpdateStrategy.Type)},
						},
					},
				}
			}),
		),

		*generator.NewFamilyGenerator(
			"kruise_daemonset_status_current_number_scheduled",
			"The number of nodes running at least one daemon pod and are supposed to.",
			metric.Gauge,
			"",
			wrapDaemonSetFunc(func(ds *v1alpha1.DaemonSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(ds.Status.CurrentNumberScheduled),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_daemonset_status_desired_number_scheduled",
			"The number of nodes that should be running the daemon pod.",
			metric.Gauge,
			"",
			wrapDaemonSetFunc(func(ds *v1alpha1.DaemonSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(ds.Status.DesiredNumberScheduled),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_daemonset_status_number_available",
			"The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and available",
			metric.Gauge,
			"",
			wrapDaemonSetFunc(func(ds *v1alpha1.DaemonSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(ds.Status.NumberAvailable),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_daemonset_status_number_misscheduled",
			"The number of nodes running a daemon pod but are not supposed to.",
			metric.Gauge,
			"",
			wrapDaemonSetFunc(func(ds *v1alpha1.DaemonSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(ds.Status.NumberMisscheduled),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_daemonset_status_number_ready",
			"The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and ready.",
			metric.Gauge,
			"",
			wrapDaemonSetFunc(func(ds *v1alpha1.DaemonSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(ds.Status.NumberReady),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_daemonset_status_number_unavailable",
			"The number of nodes that should be running the daemon pod and have none of the daemon pod running and available",
			metric.Gauge,
			"",
			wrapDaemonSetFunc(func(ds *v1alpha1.DaemonSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(ds.Status.NumberUnavailable),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_daemonset_status_observed_generation",
			"The most recent generation observed by the daemon set controller.",
			metric.Gauge,
			"",
			wrapDaemonSetFunc(func(ds *v1alpha1.DaemonSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(ds.Status.ObservedGeneration),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_daemonset_status_updated_number_scheduled",
			"The total number of nodes that are running updated daemon pod",
			metric.Gauge,
			"",
			wrapDaemonSetFunc(func(ds *v1alpha1.DaemonSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(ds.Status.UpdatedNumberScheduled),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_daemonset_metadata_generation",
			"Sequence number representing a specific generation of the desired state.",
			metric.Gauge,
			"",
			wrapDaemonSetFunc(func(ds *v1alpha1.DaemonSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(ds.ObjectMeta.Generation),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			descDaemonSetAnnotationsName,
			descDaemonSetAnnotationsHelp,
			metric.Gauge,
			"",
			wrapDaemonSetFunc(func(ds *v1alpha1.DaemonSet) *metric.Family {
				annotationKeys, annotationValues := createPrometheusLabelKeysValues("annotation", ds.Annotations, allowAnnotationsList)
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
			descDaemonSetLabelsName,
			descDaemonSetLabelsHelp,
			metric.Gauge,
			"",
			wrapDaemonSetFunc(func(ds *v1alpha1.DaemonSet) *metric.Family {
				labelKeys, labelValues := createPrometheusLabelKeysValues("label", ds.Labels, allowLabelsList)
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
	}
}

func wrapDaemonSetFunc(f func(*v1alpha1.DaemonSet) *metric.Family) func(interface{}) *metric.Family {
	return func(obj interface{}) *metric.Family {
		daemonset := obj.(*v1alpha1.DaemonSet)

		metricFamily := f(daemonset)

		for _, m := range metricFamily.Metrics {
			m.LabelKeys = append(descDaemonSetLabelsDefaultLabels, m.LabelKeys...)
			m.LabelValues = append([]string{daemonset.Namespace, daemonset.Name}, m.LabelValues...)
		}

		return metricFamily
	}
}

func createDaemonSetListWatch(kruiseClient kruiseclientset.Interface, ns string) cache.ListerWatcher {
	return &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return kruiseClient.AppsV1alpha1().DaemonSets(ns).List(context.TODO(), opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			return kruiseClient.AppsV1alpha1().DaemonSets(ns).Watch(context.TODO(), opts)
		},
	}
}
