/*
Copyright 2021 The Kruise Authors.

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

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/prometheus/common/version"
	klog "k8s.io/klog/v2"
	"k8s.io/kube-state-metrics/v2/pkg/options"

	"github.com/openkruise/kruise-state-metrics/pkg/app"
	localoptions "github.com/openkruise/kruise-state-metrics/pkg/options"
)

func main() {
	options.DefaultResources = localoptions.DefaultResources
	opts := options.NewOptions()
	opts.AddFlags()

	err := opts.Parse()
	if err != nil {
		klog.Fatalf("Error: %s", err)
	}

	if opts.Version {
		fmt.Printf("%s\n", version.Print("kruise-state-metrics"))
		os.Exit(0)
	}

	if opts.Help {
		opts.Usage()
		os.Exit(0)
	}

	ctx := context.Background()
	if err := app.RunKruiseStateMetrics(ctx, opts); err != nil {
		klog.Fatalf("Failed to run kruise-state-metrics: %v", err)
	}
}
