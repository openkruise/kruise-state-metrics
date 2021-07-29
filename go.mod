module github.com/openkruise/kruise-state-metrics

go 1.16

require (
	github.com/oklog/run v1.1.0
	github.com/openkruise/kruise-api v0.9.0-1.18
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.29.0
	github.com/prometheus/exporter-toolkit v0.6.0
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/autoscaler/vertical-pod-autoscaler v0.9.2
	k8s.io/client-go v0.21.2
	k8s.io/klog/v2 v2.9.0
	k8s.io/kube-state-metrics/v2 v2.1.1-0.20210714123226-1d61fc146160
)

// replace "github.com/openkruise/kruise/apis/apps/v1alpha1" => "../kruise"
