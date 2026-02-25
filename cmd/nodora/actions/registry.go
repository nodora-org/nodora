package actions

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/urfave/cli/v3"
	"nodora.org/nodora/pkg/registry"
)

func Registry(ctx context.Context, cmd *cli.Command) error {
	ns := cmd.Args().First()
	if ns == "" {
		return listNamespaces()
	}
	return describeNamespace(ns)
}

func listNamespaces() error {
	namespaces := registry.Global().Namespaces()
	sort.Strings(namespaces)

	fmt.Println("Available namespaces:")
	fmt.Println()
	for _, ns := range namespaces {
		label := ns
		if label == "" {
			label = "core"
		}
		r := registry.Global()
		nsObj, _ := r.GetNamespace(ns)
		fmt.Printf("  %-16s %-2d function(s)\n", label, len(nsObj.Funcs))
	}
	fmt.Println()
	fmt.Println("Use 'nodora registry <namespace>' to list functions in a namespace.")
	return nil
}

func describeNamespace(ns string) error {
	if ns == "core" {
		ns = ""
	}

	r := registry.Global()
	nsObj, ok := r.GetNamespace(ns)
	if !ok {
		return fmt.Errorf("namespace '%s' not found", ns)
	}

	names := make([]string, 0, len(nsObj.Funcs))
	for name := range nsObj.Funcs {
		names = append(names, name)
	}
	sort.Strings(names)

	label := ns
	if label == "" {
		label = "core"
	}
	fmt.Printf("Namespace: %s\n", label)
	fmt.Printf("Functions: %d\n", len(names))
	fmt.Println(strings.Repeat("-", 60))

	for _, name := range names {
		fn := nsObj.Funcs[name]
		fmt.Printf("%s\n", fn.Signature())
		if fn.Description != "" {
			fmt.Printf("└─ %s\n", fn.Description)
		}
		if len(fn.Args) > 0 {
			fmt.Println("\nArguments:")
			for _, arg := range fn.Args {
				req := "required"
				if !arg.Required {
					req = "optional"
				}
				desc := ""
				if arg.Description != "" {
					desc = " - " + arg.Description
				}
				fmt.Printf("  %-6s : %-16s [%s]%s\n", arg.Name, arg.Type, req, desc)
			}
		}
		fmt.Printf("\nReturns: %s\n", fn.ReturnType)
		fmt.Println(strings.Repeat("-", 60))
	}

	return nil
}
