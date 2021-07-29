/*
Copyright 2021 The Kruise Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or impliec.
See the License for the specific language governing permissions and
limitations under the License.
*/

package store

import (
	"context"

	"github.com/openkruise/kruise-api/apps/v1alpha1"
	clonesetcore "github.com/openkruise/kruise/pkg/controller/cloneset/core"
	clonesetutils "github.com/openkruise/kruise/pkg/controller/cloneset/utils"
	"github.com/openkruise/kruise/pkg/util/fieldindex"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kruiseclientset "github.com/openkruise/kruise-api/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/watch"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"k8s.io/kube-state-metrics/v2/pkg/metric"
	generator "k8s.io/kube-state-metrics/v2/pkg/metric_generator"
)

var (
	descCloneSetLabelsName          = "kruise_cloneset_labels"
	descCloneSetLabelsHelp          = "Kruise labels converted to Prometheus labels."
	descCloneSetLabelsDefaultLabels = []string{"namespace", "cloneset"}
)

func cloneSetMetricFamilies(Client clientset.Interface, allowLabelsList []string) []generator.FamilyGenerator {
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
				maxUnavailable, err := intstr.GetValueFromIntOrPercent(cs.Spec.UpdateStrategy.MaxUnavailable, int(*cs.Spec.Replicas), false)
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
				maxSurge, err := intstr.GetValueFromIntOrPercent(cs.Spec.UpdateStrategy.MaxSurge, int(*cs.Spec.Replicas), true)
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
			descCloneSetLabelsName,
			descCloneSetLabelsHelp,
			metric.Gauge,
			"",
			wrapCloneSetFunc(func(cs *v1alpha1.CloneSet) *metric.Family {
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
			"kruise_cloneset_status_replicasv_ready",
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
			"kruise_cloneset_status_replicasv_updated_ready",
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
			"kruise_cloneset_status_replicasv_unavailable",
			"The number of unavailable replicas per cloneset.",
			metric.Gauge,
			"",
			wrapCloneSetFunc(func(cs *v1alpha1.CloneSet) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{

							Value: float64(calculateUnavailableReplicas(cs, Client)),
						},
					},
				}
			}),
		),
		// TODO
		// *generator.NewFamilyGenerator(
		// 	"kruise_cloneset_spec_strategy_partition",
		// 	"Number of desired pods for a cloneset.",
		// 	metric.Gauge,
		// 	"",
		// 	wrapCloneSetFunc(func(cs *v1alpha1.CloneSet) *metric.Family {
		// 		return &metric.Family{
		// 			Metrics: []*metric.Metric{
		// 				{
		// 					Value: float64(cs.Status.UpdatedReadyReplicas),
		// 				},
		// 			},
		// 		}
		// 	}),
		// ),
		// // TODO
		// *generator.NewFamilyGenerator(
		// 	"kruise_cloneset_spec_strategy_type",
		// 	"Number of desired pods for a cloneset.",
		// 	metric.Gauge,
		// 	"",
		// 	wrapCloneSetFunc(func(cs *v1alpha1.CloneSet) *metric.Family {
		// 		return &metric.Family{
		// 			Metrics: []*metric.Metric{
		// 				{
		// 					Value: float64(cs.Status.UpdatedReadyReplicas),
		// 				},
		// 			},
		// 		}
		// 	}),
		// ),
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

func calculateUnavailableReplicas(cs *v1alpha1.CloneSet, Client clientset.Interface) int {
	UnavailableReplicas := 0

	opts := &client.ListOptions{
		Namespace:     cs.Namespace,
		FieldSelector: fields.SelectorFromSet(fields.Set{fieldindex.IndexNameForOwnerRefUID: string(cs.UID)}),
	}
	pods, err := clonesetutils.GetActivePods(Client, opts)
	if err != nil {
		return UnavailableReplicas
	}
	coreControl := clonesetcore.New(cs)
	for _, pod := range pods {
		if coreControl.IsPodUpdateReady(pod, 0) {
			if !coreControl.IsPodUpdateReady(pod, cs.Spec.MinReadySeconds) {
				UnavailableReplicas++
			}
		}

	}
	return UnavailableReplicas
}
