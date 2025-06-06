/*
Copyright 2021 The Kruise Authors.
Copyright 2018 The Kubernetes Authors All rights reserved.

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
	"testing"
)

func TestSortLabels(t *testing.T) {
	in := `kube_pod_container_info{container_id="docker://cd456",image="registry.k8s.io/hyperkube2",container="container2",image_id="docker://sha256:bbb",namespace="ns2",pod="pod2"} 1
kube_pod_container_info{namespace="ns2",container="container3",container_id="docker://ef789",image="registry.k8s.io/hyperkube3",image_id="docker://sha256:ccc",pod="pod2"} 1`

	want := `kube_pod_container_info{container="container2",container_id="docker://cd456",image="registry.k8s.io/hyperkube2",image_id="docker://sha256:bbb",namespace="ns2",pod="pod2"} 1
kube_pod_container_info{container="container3",container_id="docker://ef789",image="registry.k8s.io/hyperkube3",image_id="docker://sha256:ccc",namespace="ns2",pod="pod2"} 1`

	out := sortLabels(in)

	if want != out {
		t.Fatalf("expected:\n%v\nbut got:\n%v", want, out)
	}
}

func TestRemoveUnusedWhitespace(t *testing.T) {
	in := "       kube_cron_job_info \n        kube_pod_container_info \n        kube_config_map_info     "

	want := "kube_cron_job_info\nkube_pod_container_info\nkube_config_map_info"

	out := removeUnusedWhitespace(in)

	if want != out {
		t.Fatalf("expected: %q\nbut got: %q", want, out)
	}
}

func TestSortByLine(t *testing.T) {
	in := "kube_cron_job_info \nkube_pod_container_info \nkube_config_map_info"

	want := "kube_config_map_info\nkube_cron_job_info \nkube_pod_container_info "

	out := sortByLine(in)

	if want != out {
		t.Fatalf("expected: %q\nbut got: %q", want, out)
	}
}
