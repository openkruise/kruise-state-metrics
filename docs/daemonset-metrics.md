# DaemonSet Metrics

| Metric name| Metric type | Status |
| ---------- | ----------- | ----------- |
| kruise_daemonset_created | Unix creation timestamp | STABLE |
| kruise_daemonset_status_condition | The current status conditions of a daemonset | STABLE |
| kruise_daemonset_spec_strategy_rollingupdate_max_surge | Maximum number of replicas that can be scheduled above the desired number of replicas during a rolling update of a daemonset | STABLE |
| kruise_daemonset_spec_strategy_partition | Desired number or percent of Pods in old revisions | STABLE |
| kruise_daemonset_spec_strategy_type | The type of updateStrategy | STABLE |
| kruise_daemonset_status_current_number_scheduled | The number of nodes running at least one daemon pod and are supposed to | STABLE |
| kruise_daemonset_status_desired_number_scheduled | The number of nodes that should be running the daemon pod | STABLE |
| kruise_daemonset_status_number_available | The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and available | STABLE |
| kruise_daemonset_status_number_misscheduled | The number of nodes running a daemon pod but are not supposed to | STABLE |
| kruise_daemonset_status_number_ready | The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and ready | STABLE |
| kruise_daemonset_status_number_unavailable | The number of nodes that should be running the daemon pod and have none of the daemon pod running and available | STABLE |
| kruise_daemonset_status_observed_generation | The most recent generation observed by the daemon set controller | STABLE |
| kruise_daemonset_status_updated_number_scheduled | The total number of nodes that are running updated daemon pod | STABLE |
| kruise_daemonset_metadata_generation | Sequence number representing a specific generation of the desired state | STABLE |
| kruise_daemonset_labels | Kruise labels converted to Prometheus labels | STABLE |