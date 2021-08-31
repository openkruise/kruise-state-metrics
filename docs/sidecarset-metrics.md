# SidecarSet Metrics

| Metric name| Description | Status |
| ---------- | ----------- | ----------- |
| kruise_sidecarset_created | Unix creation timestamp | STABLE |
| kruise_sidecarset_status_replicas_matched | The number of matched replicas per sidecarset | STABLE |
| kruise_sidecarset_status_replicas_updated | The number of updated replicas per sidecarset | STABLE |
| kruise_sidecarset_status_replicas_ready | The number of ready replicas per sidecarset | STABLE |
| kruise_sidecarset_status_observed_generation | The generation observed by the sidecarset controller | STABLE |
| kruise_sidecarset_status_replicas_updated_ready | The number of update and ready replicas per sidecarset | STABLE |

| kruise_sidecarset_spec_namespcace | The namespace matched pods in | STABLE |
| kruise_sidecarset_spec_strategy_rollingupdate_max_unavailable | Maximum number of unavailable replicas during a rolling update of a sidecarset | STABLE |
| kruise_sidecarset_spec_strategy_partition | Desired number or percent of Pods in old revisions | STABLE |
| kruise_sidecarset_spec_strategy_type | The type of updateStrategy | STABLE |
| kruise_sidecarset_spec_metadata_generation | Sequence number representing a specific generation of the desired state | STABLE |
| kruise_sidecarset_labels | Kruise labels converted to Prometheus labels | STABLE |
| kruise_sidecarset_spec_containers_injectpolicy |  | STABLE |
| kruise_sidecarset_spec_containers_strategy_type |  | STABLE |
| kruise_sidecarset_spec_containers_strategy_hotupgradeemptyimage |  | STABLE |
| kruise_sidecarset_spec_containers_volumepolicy |  | STABLE |