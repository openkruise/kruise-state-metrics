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
	descWorkloadSpreadAnnotationsName     = "kruise_workloadspread_annotations"
	descWorkloadSpreadAnnotationsHelp     = "Kruise annotations converted to Prometheus labels."
	descWorkloadSpreadLabelsName          = "kruise_workloadspread_labels"
	descWorkloadSpreadLabelsHelp          = "Kruise labels converted to Prometheus labels."
	descWorkloadSpreadLabelsDefaultLabels = []string{"namespace", "workloadspread"}
)

func workloadSpreadMetricFamilies(allowAnnotationsList, allowLabelsList []string) []generator.FamilyGenerator {
	return []generator.FamilyGenerator{
		*generator.NewFamilyGenerator(
			"kruise_workloadspread_created",
			"Unix creation timestamp",
			metric.Gauge,
			"",
			wrapWorkloadSpreadFunc(func(ws *v1alpha1.WorkloadSpread) *metric.Family {
				ms := []*metric.Metric{}

				if !ws.CreationTimestamp.IsZero() {
					ms = append(ms, &metric.Metric{
						Value: float64(ws.CreationTimestamp.Unix()),
					})
				}

				return &metric.Family{
					Metrics: ms,
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_workloadspread_status_subset_replicas",
			"The most recently observed number of replicas for subset.",
			metric.Gauge,
			"",
			wrapWorkloadSpreadFunc(func(ws *v1alpha1.WorkloadSpread) *metric.Family {
				ms := []*metric.Metric{}
				for _, subset := range ws.Status.SubsetStatuses {
					ms = append(ms, &metric.Metric{
						LabelKeys:   []string{"replicas"},
						LabelValues: []string{string(subset.Replicas)},
					})
				}
				return &metric.Family{
					Metrics: ms,
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_workloadspread_status_subset_replicas_missing",
			"The number of replicas belong to this subset not be found.",
			metric.Gauge,
			"",
			wrapWorkloadSpreadFunc(func(ws *v1alpha1.WorkloadSpread) *metric.Family {
				ms := []*metric.Metric{}
				for _, subset := range ws.Status.SubsetStatuses {
					ms = append(ms, &metric.Metric{
						LabelKeys:   []string{"missingreplicas"},
						LabelValues: []string{string(subset.MissingReplicas)},
					})
				}
				return &metric.Family{
					Metrics: ms,
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_workloadspread_metadata_generation",
			"Sequence number representing a specific generation of the desired state for the workloadspread.",
			metric.Gauge,
			"",
			wrapWorkloadSpreadFunc(func(ws *v1alpha1.WorkloadSpread) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(ws.ObjectMeta.Generation),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			descWorkloadSpreadAnnotationsName,
			descWorkloadSpreadAnnotationsHelp,
			metric.Gauge,
			"",
			wrapWorkloadSpreadFunc(func(ws *v1alpha1.WorkloadSpread) *metric.Family {
				annotationKeys, annotationValues := createPrometheusLabelKeysValues("annotation", ws.Annotations, allowAnnotationsList)
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
			descWorkloadSpreadLabelsName,
			descWorkloadSpreadLabelsHelp,
			metric.Gauge,
			"",
			wrapWorkloadSpreadFunc(func(ws *v1alpha1.WorkloadSpread) *metric.Family {
				labelKeys, labelValues := createPrometheusLabelKeysValues("label", ws.Labels, allowLabelsList)
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
			"kruise_workloadspread_spec_strategy_type",
			"The type of updateStrategy.",
			metric.Gauge,
			"",
			wrapWorkloadSpreadFunc(func(ws *v1alpha1.WorkloadSpread) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							LabelKeys:   []string{"strategy_type"},
							LabelValues: []string{string(ws.Spec.ScheduleStrategy.Type)},
						},
					},
				}
			}),
		),
	}
}

func wrapWorkloadSpreadFunc(f func(*v1alpha1.WorkloadSpread) *metric.Family) func(interface{}) *metric.Family {
	return func(obj interface{}) *metric.Family {
		workloadspread := obj.(*v1alpha1.WorkloadSpread)

		metricFamily := f(workloadspread)

		for _, m := range metricFamily.Metrics {
			m.LabelKeys = append(descWorkloadSpreadLabelsDefaultLabels, m.LabelKeys...)
			m.LabelValues = append([]string{workloadspread.Namespace, workloadspread.Name}, m.LabelValues...)
		}

		return metricFamily
	}
}

func createWorkloadSpreadListWatch(kruiseClient kruiseclientset.Interface, ns string) cache.ListerWatcher {
	return &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return kruiseClient.AppsV1alpha1().WorkloadSpreads(ns).List(context.TODO(), opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			return kruiseClient.AppsV1alpha1().WorkloadSpreads(ns).Watch(context.TODO(), opts)
		},
	}
}
