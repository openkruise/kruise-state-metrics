# CloneSet Metrics

| Metric name| Metric type | Labels/tags | Status |
| ---------- | ----------- | ----------- | ----------- |
| kruise_cloneset_status_replicas | Gauge | `cloneset`=&lt;cloneset-name&gt; <br> `namespace`=&lt;cloneset-namespace&gt; | DEVELOP |
| kruise_cloneset_status_replicas_available | Gauge | `cloneset`=&lt;cloneset-name&gt; <br> `namespace`=&lt;cloneset-namespace&gt; | DEVELOP |
| kruise_cloneset_status_replicas_unavailable | Gauge | `cloneset`=&lt;cloneset-name&gt; <br> `namespace`=&lt;cloneset-namespace&gt; | DEVELOP |
| kruise_cloneset_status_replicas_updated | Gauge | `cloneset`=&lt;cloneset-name&gt; <br> `namespace`=&lt;cloneset-namespace&gt; | DEVELOP |
| kruise_cloneset_status_observed_generation | Gauge | `cloneset`=&lt;cloneset-name&gt; <br> `namespace`=&lt;cloneset-namespace&gt; | DEVELOP |
| kruise_cloneset_status_condition | Gauge | `cloneset`=&lt;cloneset-name&gt; <br> `namespace`=&lt;cloneset-namespace&gt; <br> `condition`=&lt;cloneset-condition&gt; <br> `status`=&lt;true\|false\|unknown&gt; | DEVELOP |
| kruise_cloneset_spec_replicas | Gauge | `cloneset`=&lt;cloneset-name&gt; <br> `namespace`=&lt;cloneset-namespace&gt; | DEVELOP |
| kruise_cloneset_spec_paused | Gauge | `cloneset`=&lt;cloneset-name&gt; <br> `namespace`=&lt;cloneset-namespace&gt; | DEVELOP |
| kruise_cloneset_spec_strategy_rollingupdate_max_unavailable | Gauge | `cloneset`=&lt;cloneset-name&gt; <br> `namespace`=&lt;cloneset-namespace&gt; | DEVELOP |
| kruise_cloneset_spec_strategy_rollingupdate_max_surge | Gauge | `cloneset`=&lt;cloneset-name&gt; <br> `namespace`=&lt;cloneset-namespace&gt; | DEVELOP |
| kruise_cloneset_metadata_generation | Gauge | `cloneset`=&lt;cloneset-name&gt; <br> `namespace`=&lt;cloneset-namespace&gt; | DEVELOP |
| kruise_cloneset_labels | Gauge | `cloneset`=&lt;cloneset-name&gt; <br> `namespace`=&lt;cloneset-namespace&gt; <br> `label_cloneset_LABEL`=&lt;cloneset_LABEL&gt; | DEVELOP |
| kruise_cloneset_created | Gauge | `cloneset`=&lt;cloneset-name&gt; <br> `namespace`=&lt;cloneset-namespace&gt; | DEVELOP |
