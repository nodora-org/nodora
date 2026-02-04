package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
	"nodora.org/nodora/cmd/nodora/actions"
)

func main() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:  "compile",
				Usage: "Compile a rule file",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "file", Aliases: []string{"f"}, Required: true},
					&cli.StringFlag{Name: "output", Aliases: []string{"o"}},
				},
				Action: actions.Compile,
			},
			{
				Name:    "eval",
				Aliases: []string{"run"},
				Usage:   "Evaluate a rule",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "file", Aliases: []string{"f"}, Required: true},
					&cli.StringFlag{Name: "rule", Aliases: []string{"r"}},
					&cli.StringFlag{Name: "input-file", Aliases: []string{"i"}},
					&cli.BoolFlag{Name: "stdin"},
					&cli.BoolFlag{Name: "debug"},
				},
				Action: actions.Eval,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Printf("💀 %v\n", err)
	}
}
