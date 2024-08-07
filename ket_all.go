/*
Copyright 2019 Cornelius Weig

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

package ketall

import (
	"context"
	"io"
	"text/tabwriter"

	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/klog/v2"

	"github.com/SkYNewZ/ketall/pkg/client"
	"github.com/SkYNewZ/ketall/pkg/filter"
	"github.com/SkYNewZ/ketall/pkg/options"
	"github.com/SkYNewZ/ketall/pkg/printer"
)

func KetAll(ctx context.Context, ketallOptions *options.KetallOptions) {
	all, err := client.GetAllServerResources(ctx, ketallOptions)
	if err != nil {
		klog.Fatal(err)
	}

	filtered := filter.ApplyFilter(ketallOptions, all)

	out := ketallOptions.Streams.Out
	if filtered == nil {
		_, _ = io.WriteString(out, "No resources found.\n")
		return
	}

	resourcePrinter, err := ketallOptions.PrintFlags.ToPrinter()
	if err != nil {
		klog.Fatal(err)
	}

	p := resourcePrinter
	switch pr := resourcePrinter.(type) {
	// yaml and json printers should operate on the full tree structure with nested lists
	case *printers.JSONPrinter:
		p = printer.NewListAdapterPrinter(pr)
	case *printers.YAMLPrinter:
		p = printer.NewListAdapterPrinter(pr)
	// other printers should flatten the resource list and operate on leaf items
	case *printer.TablePrinter:
		klog.V(2).Info("Using tabwriter")
		tw := tabwriter.NewWriter(out, 4, 4, 2, ' ', 0)
		defer tw.Flush()
		out = tw
		if err := pr.PrintHeader(out); err != nil {
			klog.Fatal("print header", err)
		}
		p = printer.NewFlattenListAdapterPrinter(pr)
	default:
		p = printer.NewFlattenListAdapterPrinter(pr)
	}

	if err = p.PrintObj(filtered, out); err != nil {
		klog.Fatal(err)
	}
}
