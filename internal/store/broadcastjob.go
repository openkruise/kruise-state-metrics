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
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"k8s.io/kube-state-metrics/v2/pkg/metric"
	generator "k8s.io/kube-state-metrics/v2/pkg/metric_generator"
)

var (
	descBroadcastJobAnnotationsName     = "kruise_broadcastjob_annotations"
	descBroadcastJobAnnotationsHelp     = "Kruise annotations converted to Prometheus labels."
	descBroadcastJobLabelsName          = "kruise_broadcastjob_labels"
	descBroadcastJobLabelsHelp          = "Kruise labels converted to Prometheus labels."
	descBroadcastJobLabelsDefaultLabels = []string{"namespace", "broadcastjob"}
)

func broadcastJobMetricFamilies(allowAnnotationsList, allowLabelsList []string) []generator.FamilyGenerator {
	return []generator.FamilyGenerator{
		*generator.NewFamilyGenerator(
			"kruise_broadcastjob_created",
			"Unix creation timestamp",
			metric.Gauge,
			"",
			wrapBroadcastJobFunc(func(bj *v1alpha1.BroadcastJob) *metric.Family {
				ms := []*metric.Metric{}

				if !bj.CreationTimestamp.IsZero() {
					ms = append(ms, &metric.Metric{
						Value: float64(bj.CreationTimestamp.Unix()),
					})
				}

				return &metric.Family{
					Metrics: ms,
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_broadcastjob_status_active",
			"The number of actively running pods.",
			metric.Gauge,
			"",
			wrapBroadcastJobFunc(func(bj *v1alpha1.BroadcastJob) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(bj.Status.Active),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_broadcastjob_status_succeeded",
			"The number of pods which reached phase Succeeded.",
			metric.Gauge,
			"",
			wrapBroadcastJobFunc(func(bj *v1alpha1.BroadcastJob) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(bj.Status.Succeeded),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_broadcastjob_status_failed",
			"The number of pods which reached phase Failed.",
			metric.Gauge,
			"",
			wrapBroadcastJobFunc(func(bj *v1alpha1.BroadcastJob) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(bj.Status.Failed),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_broadcastjob_status_desired",
			"The desired number of pods, this is typically equal to the number of nodes satisfied to run pods.",
			metric.Gauge,
			"",
			wrapBroadcastJobFunc(func(bj *v1alpha1.BroadcastJob) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(bj.Status.Desired),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_broadcastjob_status_condition",
			"The current status conditions of a broadcastjob.",
			metric.Gauge,
			"",
			wrapBroadcastJobFunc(func(bj *v1alpha1.BroadcastJob) *metric.Family {
				ms := make([]*metric.Metric, len(bj.Status.Conditions)*len(conditionStatuses))

				for i, bj := range bj.Status.Conditions {
					conditionMetrics := addConditionMetrics(bj.Status)

					for j, m := range conditionMetrics {
						metric := m

						metric.LabelKeys = []string{"condition", "status"}
						metric.LabelValues = append([]string{string(bj.Type)}, metric.LabelValues...)
						ms[i*len(conditionStatuses)+j] = metric
					}
				}

				return &metric.Family{
					Metrics: ms,
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_broadcastjob_spec_parallelism",
			"The maximum desired number of pods the job should.",
			metric.Gauge,
			"",
			wrapBroadcastJobFunc(func(bj *v1alpha1.BroadcastJob) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(bj.Spec.Parallelism.IntVal),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_broadcastjob_metadata_generation",
			"Sequence number representing a specific generation of the desired state.",
			metric.Gauge,
			"",
			wrapBroadcastJobFunc(func(bj *v1alpha1.BroadcastJob) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(bj.ObjectMeta.Generation),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			descBroadcastJobAnnotationsName,
			descBroadcastJobAnnotationsHelp,
			metric.Gauge,
			"",
			wrapBroadcastJobFunc(func(bj *v1alpha1.BroadcastJob) *metric.Family {
				annotationKeys, annotationValues := createPrometheusLabelKeysValues("annotation", bj.Annotations, allowAnnotationsList)
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
			descBroadcastJobLabelsName,
			descBroadcastJobLabelsHelp,
			metric.Gauge,
			"",
			wrapBroadcastJobFunc(func(bj *v1alpha1.BroadcastJob) *metric.Family {
				labelKeys, labelValues := createPrometheusLabelKeysValues("label", bj.Labels, allowLabelsList)
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
			"kruise_broadcastjob_spec_strategy_activedeadline_seconds",
			"The duration in seconds relative to the startTime that the job may be active.",
			metric.Gauge,
			"",
			wrapBroadcastJobFunc(func(bj *v1alpha1.BroadcastJob) *metric.Family {
				value := 0.0
				if bj != nil && bj.Spec.CompletionPolicy.ActiveDeadlineSeconds != nil {
					value = float64(*bj.Spec.CompletionPolicy.ActiveDeadlineSeconds)
				}
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: value,
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_broadcastjob_spec_strategy_ttl_seconds",
			"The lifetime of a Job that has finished.",
			metric.Gauge,
			"",
			wrapBroadcastJobFunc(func(bj *v1alpha1.BroadcastJob) *metric.Family {
				value := 0.0
				if bj != nil && bj.Spec.CompletionPolicy.TTLSecondsAfterFinished != nil {
					value = float64(*bj.Spec.CompletionPolicy.TTLSecondsAfterFinished)
				}
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: value,
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_broadcastjob_spec_strategy_type",
			"The type of completionpolicy.",
			metric.Gauge,
			"",
			wrapBroadcastJobFunc(func(bj *v1alpha1.BroadcastJob) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							LabelKeys:   []string{"strategy_type"},
							LabelValues: []string{string(bj.Spec.CompletionPolicy.Type)},
						},
					},
				}
			}),
		),
	}
}

func wrapBroadcastJobFunc(f func(*v1alpha1.BroadcastJob) *metric.Family) func(interface{}) *metric.Family {
	return func(obj interface{}) *metric.Family {
		broadcastjob := obj.(*v1alpha1.BroadcastJob)

		metricFamily := f(broadcastjob)

		for _, m := range metricFamily.Metrics {
			m.LabelKeys = append(descBroadcastJobLabelsDefaultLabels, m.LabelKeys...)
			m.LabelValues = append([]string{broadcastjob.Namespace, broadcastjob.Name}, m.LabelValues...)
		}

		return metricFamily
	}
}

func createBroadcastJobListWatch(kruiseClient kruiseclientset.Interface, ns string) cache.ListerWatcher {
	return &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return kruiseClient.AppsV1alpha1().BroadcastJobs(ns).List(context.TODO(), opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			return kruiseClient.AppsV1alpha1().BroadcastJobs(ns).Watch(context.TODO(), opts)
		},
	}
}
