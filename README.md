# kruise-state-metrics

kruise-state-metrics is a simple service that listens to the Kubernetes API
server and generates metrics about the state of the objects. (See examples in
the Metrics section below.) It is not focused on the health of the individual
OpenKruise components, but rather on the health of the various objects inside,
such as clonesets, advanced statefulsets and sidecarsets.

kruise-state-metrics is about generating metrics from OpenKruise API objects
without modification. This ensures that features provided by kruise-state-metrics
have the same grade of stability as the OpenKruise API objects themselves. In
turn, this means that kruise-state-metrics in certain situations may not show the
exact same values as kubectl, as kubectl applies certain heuristics to display
comprehensible messages. kruise-state-metrics exposes raw data unmodified from the
Kubernetes API, this way users have all the data they require and perform
heuristics as they see fit.

The metrics are exported on the HTTP endpoint `/metrics` on the listening port
(default 8080). They are served as plaintext. They are designed to be consumed
either by Prometheus itself or by a scraper that is compatible with scraping a
Prometheus client endpoint. You can also open `/metrics` in a browser to see
the raw metrics. Note that the metrics exposed on the `/metrics` endpoint
reflect the current state of OpenKruise objects in the Kubernetes cluster.
When the objects are deleted they are no longer visible on the `/metrics` endpoint.

# Installation

## Install with helm

kruise-state-metrics can be simply installed by helm v3.5+, which is a simple command-line tool and you can get it from [here](https://github.com/helm/helm/releases).

```bash
# Firstly add openkruise charts repository if you haven't do this.
$ helm repo add openkruise https://openkruise.github.io/charts/

# [Optional]
$ helm repo update

# Install the latest version.
$ helm install openkruise/kruise-state-metrics --version 0.1.0
```

## Optional: download charts manually

If you have problem with connecting to `https://openkruise.github.io/charts/` in production, you might need to download the chart from [here](https://github.com/openkruise/charts/releases) manually and install or upgrade with it.

```bash
$ helm install/upgrade kruise-state-metrics /PATH/TO/CHART
```

# Metrics Documentation
See the [`docs`](docs) directory for more information on the exposed metrics.

# Usage
See the [`examples`](examples) directory for example grafana dashboard that use metrics exposed by kruise-state-metrics


