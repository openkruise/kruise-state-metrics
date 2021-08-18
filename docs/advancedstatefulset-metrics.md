# AdvancedStateful Set Metrics

| Metric name| Description | Status |
| ---------- | ----------- | ----------- |
| kruise_advancedstatefulset_created | Unix creation timestamp | STABLE |
| kruise_advancedstatefulset_status_replicas | The number of replicas per cloneset | STABLE |
| kruise_advancedstatefulset_status_replicas_available | The number of available replicas per StatefulSet. | STABLE |
| kruise_advancedstatefulset_status_replicas_current | The number of current replicas per StatefulSet. | STABLE |
| kruise_advancedstatefulset_status_replicas_ready | The number of ready replicas per StatefulSet. | STABLE |
| kruise_advancedstatefulset_status_replicas_updated | The number of updated replicas per StatefulSet. | STABLE |
| kruise_advancedstatefulset_status_observed_generation | The generation observed by the StatefulSet controller. | STABLE |
| kruise_advancedstatefulset_status_condition | The current status conditions of a advancedstatefulset | STABLE |
| kruise_advancedstatefulset_replicas | Number of desired pods for a StatefulSet. | STABLE |
| kruise_advancedstatefulset_metadata_generation | Sequence number representing a specific generation of the desired state for the StatefulSet. | STABLE |
| kruise_advancedstatefulset_spec_replicas | Number of desired pods for a advancedstatefulset | STABLE |
| kruise_advancedstatefulset_spec_strategy_rollingupdate_max_unavailable | Maximum number of unavailable replicas during a rolling update of a advancedstatefulset | STABLE |
| kruise_advancedstatefulset_spec_reserveordinals |  | STABLE |
| kruise_advancedstatefulset_labels | Kubernetes labels converted to Prometheus labels. | STABLE |
| kruise_advancedstatefulset_status_current_revision | Indicates the version of the StatefulSet used to generate Pods in the sequence [0,currentReplicas). | STABLE |
| kruise_advancedstatefulset_status_update_revision | Indicates the version of the StatefulSet used to generate Pods in the sequence [replicas-updatedReplicas,replicas) | STABLE |
| kruise_advancedstatefulset_strategy_type | The type of updateStrategy | STABLE |