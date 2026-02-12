package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
	"nodora.org/nodora/internal/semantics"
	"nodora.org/nodora/pkg/compiler"
)

func Compile(ctx context.Context, cmd *cli.Command) error {
	filePath := cmd.String("file")

	var input []byte
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("error reading file: %v", err)
		}
		input = data
	}

	c := compiler.NewCompiler()
	nir, err := c.Compile(string(input))

	if err != nil {
		semErrs, ok := err.(*semantics.SemanticErrors)
		if !ok {
			return fmt.Errorf("%v:%w", filePath, err)
		}

		errCount := semErrs.Count()
		if errCount > 0 {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Found %v issues in %v\n\n", errCount, filePath))
			for _, e := range semErrs.Errors {
				sb.WriteString(fmt.Sprintf("> %v:%v\n", filePath, e))
			}
			return fmt.Errorf("%v", sb.String())
		}
	}

	nirEncoded, err := json.Marshal(nir)
	if err != nil {
		return fmt.Errorf("json encoding failed: %v", err)
	}

	outputPath := cmd.String("output")
	if outputPath == "" {
		outputPath = strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".json"
	}

	if err := os.WriteFile(outputPath, nirEncoded, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	fmt.Printf("\\(^_^)/ successfully compiled %s to %s\n", filePath, outputPath)
	return nil
}
