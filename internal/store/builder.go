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
	"reflect"
	"sort"
	"strconv"
	"strings"

	appsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	appsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	kruiseclientset "github.com/openkruise/kruise-api/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/api/core/v1"
	vpaclientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	ksmtypes "k8s.io/kube-state-metrics/v2/pkg/builder/types"
	generator "k8s.io/kube-state-metrics/v2/pkg/metric_generator"
	metricsstore "k8s.io/kube-state-metrics/v2/pkg/metrics_store"
	"k8s.io/kube-state-metrics/v2/pkg/options"
	"k8s.io/kube-state-metrics/v2/pkg/sharding"
	"k8s.io/kube-state-metrics/v2/pkg/watch"
)

// BuildKruiseStoresFunc function signature that is used to return a list of metricsstore.MetricsStore
type BuildKruiseStoresFunc func(metricFamilies []generator.FamilyGenerator,
	expectedType interface{},
	listWatchFunc func(kruiseClient kruiseclientset.Interface, ns string) cache.ListerWatcher,
	useAPIServerCache bool,
) []*metricsstore.MetricsStore

// Make sure the internal Builder implements the public BuilderInterface.
// New Builder methods should be added to the public BuilderInterface.
var _ ksmtypes.BuilderInterface = &Builder{}

// Builder helps to build store. It follows the builder pattern
// (https://en.wikipedia.org/wiki/Builder_pattern).
type Builder struct {
	kubeClient            clientset.Interface
	kruiseClient          kruiseclientset.Interface
	namespaces            options.NamespaceList
	ctx                   context.Context
	enabledResources      []string
	allowDenyList         ksmtypes.AllowDenyLister
	listWatchMetrics      *watch.ListWatchMetrics
	shardingMetrics       *sharding.Metrics
	shard                 int32
	totalShards           int
	buildStoresFunc       ksmtypes.BuildStoresFunc
	buildKruiseStoresFunc BuildKruiseStoresFunc
	allowAnnotationsList  map[string][]string
	allowLabelsList       map[string][]string
	useAPIServerCache     bool
}

// NewBuilder returns a new builder.
func NewBuilder() *Builder {
	b := &Builder{}
	return b
}

// WithMetrics sets the metrics property of a Builder.
func (b *Builder) WithMetrics(r prometheus.Registerer) {
	b.listWatchMetrics = watch.NewListWatchMetrics(r)
	b.shardingMetrics = sharding.NewShardingMetrics(r)
}

// WithEnabledResources sets the enabledResources property of a Builder.
func (b *Builder) WithEnabledResources(r []string) error {
	for _, col := range r {
		if !resourceExists(col) {
			return errors.Errorf("resource %s does not exist. Available resources: %s", col, strings.Join(availableResources(), ","))
		}
	}

	var copy []string
	copy = append(copy, r...)

	sort.Strings(copy)

	b.enabledResources = copy
	return nil
}

// WithNamespaces sets the namespaces property of a Builder.
func (b *Builder) WithNamespaces(n options.NamespaceList) {
	b.namespaces = n
}

// WithSharding sets the shard and totalShards property of a Builder.
func (b *Builder) WithSharding(shard int32, totalShards int) {
	b.shard = shard
	labels := map[string]string{sharding.LabelOrdinal: strconv.Itoa(int(shard))}
	b.shardingMetrics.Ordinal.Reset()
	b.shardingMetrics.Ordinal.With(labels).Set(float64(shard))
	b.totalShards = totalShards
	b.shardingMetrics.Total.Set(float64(totalShards))
}

// WithContext sets the ctx property of a Builder.
func (b *Builder) WithContext(ctx context.Context) {
	b.ctx = ctx
}

// WithKubeClient sets the kubeClient property of a Builder.
func (b *Builder) WithKubeClient(c clientset.Interface) {
	b.kubeClient = c
}

// WithKruiseClient sets the kruiseClient property of a Builder.
func (b *Builder) WithKruiseClient(c kruiseclientset.Interface) {
	b.kruiseClient = c
}

// WithVPAClient sets the vpaClient property of a Builder so that the verticalpodautoscaler collector can query VPA objects.
func (b *Builder) WithVPAClient(c vpaclientset.Interface) {
	// nothing to do
}

// WithAllowDenyList configures the allow or denylisted metric to be exposed
// by the store build by the Builder.
func (b *Builder) WithAllowDenyList(l ksmtypes.AllowDenyLister) {
	b.allowDenyList = l
}

// WithGenerateStoresFunc configures a custom generate store function
func (b *Builder) WithGenerateStoresFunc(f ksmtypes.BuildStoresFunc, u bool) {
	b.buildStoresFunc = f
	b.useAPIServerCache = u
}

// DefaultGenerateStoresFunc returns default buildStores function
func (b *Builder) DefaultGenerateStoresFunc() ksmtypes.BuildStoresFunc {
	return b.buildStores
}

// WithAllowAnnotations configures which annotations can be returned for metrics
func (b *Builder) WithAllowAnnotations(annotations map[string][]string) {
	if len(annotations) > 0 {
		b.allowAnnotationsList = annotations
	}
}

// WithAllowLabels configures which labels can be returned for metrics
func (b *Builder) WithAllowLabels(labels map[string][]string) {
	if len(labels) > 0 {
		b.allowLabelsList = labels
	}
}

// Build initializes and registers all enabled stores.
// It returns metrics writers which can be used to write out
// metrics from the stores.
func (b *Builder) Build() []metricsstore.MetricsWriter {
	if b.allowDenyList == nil {
		panic("allowDenyList should not be nil")
	}

	var metricsWriters []metricsstore.MetricsWriter
	var activeStoreNames []string

	for _, c := range b.enabledResources {
		constructor, ok := availableStores[c]
		if ok {
			stores := constructor(b)
			activeStoreNames = append(activeStoreNames, c)
			if len(stores) == 1 {
				metricsWriters = append(metricsWriters, stores[0])
			} else {
				metricsWriters = append(metricsWriters, metricsstore.NewMultiStoreMetricsWriter(stores))
			}
		}
	}

	klog.Infof("Active resources: %s", strings.Join(activeStoreNames, ","))

	return metricsWriters
}

var availableStores = map[string]func(f *Builder) []*metricsstore.MetricsStore{
	"clonesets":       func(b *Builder) []*metricsstore.MetricsStore { return b.buildCloneSetStores() },
	"statefulsets":    func(b *Builder) []*metricsstore.MetricsStore { return b.buildStatefulSetStores() },
	"sidecarsets":     func(b *Builder) []*metricsstore.MetricsStore { return b.buildSidecarSetStores() },
	"workloadspreads": func(b *Builder) []*metricsstore.MetricsStore { return b.buildWorkloadSpreadStores() },
	"daemonsets":      func(b *Builder) []*metricsstore.MetricsStore { return b.buildDaemonSetStores() },
	"broadcastjobs":   func(b *Builder) []*metricsstore.MetricsStore { return b.buildBroadcastJob() },
}

func resourceExists(name string) bool {
	_, ok := availableStores[name]
	return ok
}

func availableResources() []string {
	c := []string{}
	for name := range availableStores {
		c = append(c, name)
	}
	return c
}

// WithKruiseStoresFunc configures a custom Kruise store function
func (b *Builder) WithKruiseStoresFunc(f BuildKruiseStoresFunc, u bool) {
	b.buildKruiseStoresFunc = f
	b.useAPIServerCache = u
}

// DefaultKruiseStoresFunc returns default buildStores function
func (b *Builder) DefaultKruiseStoresFunc() BuildKruiseStoresFunc {
	return b.buildKruiseStores
}

func (b *Builder) buildCloneSetStores() []*metricsstore.MetricsStore {
	return b.buildKruiseStoresFunc(cloneSetMetricFamilies(b.allowAnnotationsList["clonesets"], b.allowLabelsList["clonesets"]), &appsv1alpha1.CloneSet{}, createCloneSetListWatch, b.useAPIServerCache)
}

func (b *Builder) buildStatefulSetStores() []*metricsstore.MetricsStore {
	return b.buildKruiseStoresFunc(statefulSetMetricFamilies(b.allowAnnotationsList["statefulsets"], b.allowLabelsList["statefulsets"]), &appsv1beta1.StatefulSet{}, createStatefulSetListWatch, b.useAPIServerCache)
}

func (b *Builder) buildSidecarSetStores() []*metricsstore.MetricsStore {
	return b.buildKruiseStoresFunc(sidecarSetMetricFamilies(b.allowAnnotationsList["sidecarsets"], b.allowLabelsList["sidecarsets"]), &appsv1alpha1.SidecarSet{}, createSidecarSetListWatch, b.useAPIServerCache)
}

func (b *Builder) buildWorkloadSpreadStores() []*metricsstore.MetricsStore {
	return b.buildKruiseStoresFunc(workloadSpreadMetricFamilies(b.allowAnnotationsList["workloadspreads"], b.allowLabelsList["workloadspreads"]), &appsv1alpha1.WorkloadSpread{}, createWorkloadSpreadListWatch, b.useAPIServerCache)
}

func (b *Builder) buildDaemonSetStores() []*metricsstore.MetricsStore {
	return b.buildKruiseStoresFunc(daemonSetMetricFamilies(b.allowAnnotationsList["daemonsets"], b.allowLabelsList["daemonsets"]), &appsv1alpha1.DaemonSet{}, createDaemonSetListWatch, b.useAPIServerCache)
}

func (b *Builder) buildBroadcastJob() []*metricsstore.MetricsStore {
	return b.buildKruiseStoresFunc(broadcastJobMetricFamilies(b.allowAnnotationsList["broadcastjobs"], b.allowLabelsList["broadcastjobs"]), &appsv1alpha1.BroadcastJob{}, createBroadcastJobListWatch, b.useAPIServerCache)
}

func (b *Builder) buildKruiseStores(
	metricFamilies []generator.FamilyGenerator,
	expectedType interface{},
	listWatchFunc func(kruiseClient kruiseclientset.Interface, ns string) cache.ListerWatcher,
	useAPIServerCache bool,
) []*metricsstore.MetricsStore {
	metricFamilies = generator.FilterMetricFamilies(b.allowDenyList, metricFamilies)
	composedMetricGenFuncs := generator.ComposeMetricGenFuncs(metricFamilies)
	familyHeaders := generator.ExtractMetricFamilyHeaders(metricFamilies)

	if isAllNamespaces(b.namespaces) {
		store := metricsstore.NewMetricsStore(
			familyHeaders,
			composedMetricGenFuncs,
		)
		listWatcher := listWatchFunc(b.kruiseClient, v1.NamespaceAll)
		b.startReflector(expectedType, store, listWatcher, useAPIServerCache)
		return []*metricsstore.MetricsStore{store}
	}

	stores := make([]*metricsstore.MetricsStore, 0, len(b.namespaces))
	for _, ns := range b.namespaces {
		store := metricsstore.NewMetricsStore(
			familyHeaders,
			composedMetricGenFuncs,
		)
		listWatcher := listWatchFunc(b.kruiseClient, ns)
		b.startReflector(expectedType, store, listWatcher, useAPIServerCache)
		stores = append(stores, store)
	}

	return stores
}

func (b *Builder) buildStores(
	metricFamilies []generator.FamilyGenerator,
	expectedType interface{},
	listWatchFunc func(kubeClient clientset.Interface, ns string) cache.ListerWatcher,
	useAPIServerCache bool,
) []*metricsstore.MetricsStore {
	metricFamilies = generator.FilterMetricFamilies(b.allowDenyList, metricFamilies)
	composedMetricGenFuncs := generator.ComposeMetricGenFuncs(metricFamilies)
	familyHeaders := generator.ExtractMetricFamilyHeaders(metricFamilies)

	if isAllNamespaces(b.namespaces) {
		store := metricsstore.NewMetricsStore(
			familyHeaders,
			composedMetricGenFuncs,
		)
		listWatcher := listWatchFunc(b.kubeClient, v1.NamespaceAll)
		b.startReflector(expectedType, store, listWatcher, useAPIServerCache)
		return []*metricsstore.MetricsStore{store}
	}

	stores := make([]*metricsstore.MetricsStore, 0, len(b.namespaces))
	for _, ns := range b.namespaces {
		store := metricsstore.NewMetricsStore(
			familyHeaders,
			composedMetricGenFuncs,
		)
		listWatcher := listWatchFunc(b.kubeClient, ns)
		b.startReflector(expectedType, store, listWatcher, useAPIServerCache)
		stores = append(stores, store)
	}

	return stores
}

// startReflector starts a Kubernetes client-go reflector with the given
// listWatcher and registers it with the given store.
func (b *Builder) startReflector(
	expectedType interface{},
	store cache.Store,
	listWatcher cache.ListerWatcher,
	useAPIServerCache bool,
) {
	instrumentedListWatch := watch.NewInstrumentedListerWatcher(listWatcher, b.listWatchMetrics, reflect.TypeOf(expectedType).String(), useAPIServerCache)
	reflector := cache.NewReflector(sharding.NewShardedListWatch(b.shard, b.totalShards, instrumentedListWatch), expectedType, store, 0)
	go reflector.Run(b.ctx.Done())
}

// isAllNamespaces checks if the given slice of namespaces
// contains only v1.NamespaceAll.
func isAllNamespaces(namespaces []string) bool {
	return len(namespaces) == 1 && namespaces[0] == v1.NamespaceAll
}
