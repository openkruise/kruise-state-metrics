# CloneSet Metrics

| Metric name| Description | Status |
| ---------- | ----------- | ----------- |
| kruise_cloneset_created | Unix creation timestamp | STABLE |
| kruise_cloneset_metadata_generation | Sequence number representing a specific generation of the desired state | STABLE |
| kruise_cloneset_status_replicas | The number of replicas per cloneset | STABLE |
| kruise_cloneset_status_replicas_available | The number of available replicas per cloneset | STABLE |
| kruise_cloneset_status_replicas_updated | The number of updated replicas per cloneset | STABLE |
| kruise_cloneset_status_observed_generation | The generation observed by the cloneset controller | STABLE |
| kruise_cloneset_status_condition | The current status conditions of a cloneset | STABLE |
| kruise_cloneset_status_replicas_ready | The number of ready replicas per cloneset | STABLE |
| kruise_cloneset_status_replicas_updated_ready | The number of update and ready replicas per cloneset | STABLE |
| kruise_cloneset_spec_replicas | Number of desired pods for a cloneset | STABLE |
| kruise_cloneset_spec_strategy_rollingupdate_max_unavailable | Maximum number of unavailable replicas during a rolling update of a cloneset | STABLE |
| kruise_cloneset_spec_strategy_rollingupdate_max_surge | Maximum number of replicas that can be scheduled above the desired number of replicas during a rolling update of a cloneset | STABLE |
| kruise_cloneset_spec_strategy_partition | Desired number or percent of Pods in old revisions | STABLE |
| kruise_cloneset_spec_strategy_type | The type of updateStrategy | STABLE |
| kruise_cloneset_labels | Kruise labels converted to Prometheus labels | STABLE |
