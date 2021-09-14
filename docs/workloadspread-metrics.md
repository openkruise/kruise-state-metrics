# Workload Spread Metrics

| Metric name| Description | Status |
| ---------- | ----------- | ----------- |
| kruise_workloadspread_created | Unix creation timestamp | STABLE |
| kruise_workloadspread_status_subset_replicas | The most recently observed number of replicas for subset. | STABLE |
| kruise_workloadspread_status_subset_replicas_missing | The number of replicas belong to this subset not be found. | STABLE |
| kruise_workloadspread_spec_metadata_generation | Sequence number representing a specific generation of the desired state for the workloadspread. | STABLE |
| kruise_workloadspread_spec_subsets_max_replicas | The desired max replicas of this subset. | STABLE |
| kruise_workloadspread_spec_strategy_type | The type of updateStrategy | STABLE |
| kruise_workloadspread_labels | Kubernetes labels converted to Prometheus labels. | STABLE |