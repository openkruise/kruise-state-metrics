/*
Copyright 2022 The Kruise Authors.

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
	descContainerRecreateRequestAnnotationsName     = "kruise_containerrecreaterequest_annotations"
	descContainerRecreateRequestAnnotationsHelp     = "Kruise annotations converted to Prometheus labels."
	descContainerRecreateRequestLabelsName          = "kruise_containerrecreaterequest_labels"
	descContainerRecreateRequestLabelsHelp          = "Kruise labels converted to Prometheus labels."
	descContainerRecreateRequestLabelsDefaultLabels = []string{"namespace", "containerrecreaterequest"}
)

func containerRecreateRequestMetricFamilies(allowAnnotationsList, allowLabelsList []string) []generator.FamilyGenerator {
	return []generator.FamilyGenerator{
		*generator.NewFamilyGenerator(
			descContainerRecreateRequestAnnotationsName,
			descContainerRecreateRequestAnnotationsHelp,
			metric.Gauge,
			"",
			wrapContainerRecreateRequestFunc(func(crr *v1alpha1.ContainerRecreateRequest) *metric.Family {
				annotationKeys, annotationValues := createPrometheusLabelKeysValues("annotation", crr.Annotations, allowAnnotationsList)
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
			descContainerRecreateRequestLabelsName,
			descContainerRecreateRequestLabelsHelp,
			metric.Gauge,
			"",
			wrapContainerRecreateRequestFunc(func(crr *v1alpha1.ContainerRecreateRequest) *metric.Family {
				labelKeys, labelValues := createPrometheusLabelKeysValues("label", crr.Labels, allowLabelsList)
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
			"kruise_containerrecreaterequest_created",
			"Unix creation timestamp",
			metric.Gauge,
			"",
			wrapContainerRecreateRequestFunc(func(crr *v1alpha1.ContainerRecreateRequest) *metric.Family {
				ms := []*metric.Metric{}

				if !crr.CreationTimestamp.IsZero() {
					ms = append(ms, &metric.Metric{
						Value: float64(crr.CreationTimestamp.Unix()),
					})
				}

				return &metric.Family{
					Metrics: ms,
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_containerrecreaterequest_containers_pending",
			"The number of containers which reached Phase Pending.",
			metric.Gauge,
			"",
			wrapContainerRecreateRequestFunc(func(crr *v1alpha1.ContainerRecreateRequest) *metric.Family {
				var count int
				for i := range crr.Status.ContainerRecreateStates {
					if crr.Status.ContainerRecreateStates[i].Phase == v1alpha1.ContainerRecreateRequestPending {
						count++
					}
				}
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(count),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_containerrecreaterequest_containers_recreating",
			"The number of containers which reached Phase Recreating.",
			metric.Gauge,
			"",
			wrapContainerRecreateRequestFunc(func(crr *v1alpha1.ContainerRecreateRequest) *metric.Family {
				var count int
				for i := range crr.Status.ContainerRecreateStates {
					if crr.Status.ContainerRecreateStates[i].Phase == v1alpha1.ContainerRecreateRequestRecreating {
						count++
					}
				}
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(count),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_containerrecreaterequest_containers_succeeded",
			"The number of containers which reached Phase Succeeded.",
			metric.Gauge,
			"",
			wrapContainerRecreateRequestFunc(func(crr *v1alpha1.ContainerRecreateRequest) *metric.Family {
				var count int
				for i := range crr.Status.ContainerRecreateStates {
					if crr.Status.ContainerRecreateStates[i].Phase == v1alpha1.ContainerRecreateRequestSucceeded {
						count++
					}
				}
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(count),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_containerrecreaterequest_containers_failed",
			"The number of containers which reached Phase Failed.",
			metric.Gauge,
			"",
			wrapContainerRecreateRequestFunc(func(crr *v1alpha1.ContainerRecreateRequest) *metric.Family {
				var count int
				for i := range crr.Status.ContainerRecreateStates {
					if crr.Status.ContainerRecreateStates[i].Phase == v1alpha1.ContainerRecreateRequestFailed {
						count++
					}
				}
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(count),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_containerrecreaterequest_pending",
			"The number of CRR which reached Phase Pending.",
			metric.Gauge,
			"",
			wrapContainerRecreateRequestFunc(func(crr *v1alpha1.ContainerRecreateRequest) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: boolFloat64(crr.Status.Phase == v1alpha1.ContainerRecreateRequestPending),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_containerrecreaterequest_recreating",
			"The number of CRR which reached Phase Recreating.",
			metric.Gauge,
			"",
			wrapContainerRecreateRequestFunc(func(crr *v1alpha1.ContainerRecreateRequest) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: boolFloat64(crr.Status.Phase == v1alpha1.ContainerRecreateRequestRecreating),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kruise_containerrecreaterequest_completed",
			"The number of CRR which reached Phase Completed.",
			metric.Gauge,
			"",
			wrapContainerRecreateRequestFunc(func(crr *v1alpha1.ContainerRecreateRequest) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: boolFloat64(crr.Status.Phase == v1alpha1.ContainerRecreateRequestCompleted),
						},
					},
				}
			}),
		),
	}
}

func wrapContainerRecreateRequestFunc(f func(*v1alpha1.ContainerRecreateRequest) *metric.Family) func(interface{}) *metric.Family {
	return func(obj interface{}) *metric.Family {
		crr := obj.(*v1alpha1.ContainerRecreateRequest)

		metricFamily := f(crr)

		for _, m := range metricFamily.Metrics {
			m.LabelKeys = append(descContainerRecreateRequestLabelsDefaultLabels, m.LabelKeys...)
			m.LabelValues = append([]string{crr.Namespace, crr.Name}, m.LabelValues...)
		}

		return metricFamily
	}
}

func createContainerRecreateRequestListWatch(kruiseClient kruiseclientset.Interface, ns string) cache.ListerWatcher {
	return &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return kruiseClient.AppsV1alpha1().ContainerRecreateRequests(ns).List(context.TODO(), opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			return kruiseClient.AppsV1alpha1().ContainerRecreateRequests(ns).Watch(context.TODO(), opts)
		},
	}
}
