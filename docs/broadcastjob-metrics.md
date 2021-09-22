# BroadcastJob Metrics

| Metric name| Description | Status |
| ---------- | ----------- | ----------- |
| kruise_broadcastjob_created | Unix creation timestamp | STABLE |
| kruise_broadcastjob_metadata_generation | Sequence number representing a specific generation of the desired state | STABLE |
| kruise_broadcastjob_status_active | The number of actively running pods | STABLE |
| kruise_broadcastjob_status_succeeded | The number of pods which reached phase Succeeded | STABLE |
| kruise_broadcastjob_status_failed | The number of pods which reached phase Failed | STABLE |
| kruise_broadcastjob_status_desired | The desired number of pods, this is typically equal to the number of nodes satisfied to run pods | STABLE |
| kruise_broadcastjob_status_condition | The current status conditions of a broadcastjob | STABLE |
| kruise_broadcastjob_status_spec_parallelism | The maximum desired number of pods the job should | STABLE |
| kruise_broadcastjob_spec_strategy_activedeadline_seconds | The duration in seconds relative to the startTime that the job may be active | STABLE |
| kruise_broadcastjob_spec_strategy_ttl_seconds | The lifetime of a Job that has finished | STABLE |
| kruise_broadcastjob_spec_strategy_type | The type of updateStrategy | STABLE |
| kruise_broadcastjob_labels | Kruise labels converted to Prometheus labels | STABLE |