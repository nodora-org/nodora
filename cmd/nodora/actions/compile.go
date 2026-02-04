package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
	"nodora.org/nodora/internal/nir"
	"nodora.org/nodora/internal/parser"
	"nodora.org/nodora/internal/semantics"
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

	p, err := parser.Parse(string(input))
	if err != nil {
		return fmt.Errorf("%v:%w", filePath, err)
	}

	analyzer := semantics.NewSemanticAnalyzer()
	errors := analyzer.Analyze(p)

	errorsCount := len(errors)
	if errorsCount > 0 {
		fmt.Printf("Found %v issues in %v\n", errorsCount, filePath)

		for _, err := range errors {
			fmt.Printf("\t%v\n", err)
		}

		return nil
	}

	converter := nir.NewConverter()
	nirProg, err := converter.ConvertFromAST(p)
	if err != nil {
		return fmt.Errorf("conversion failed: %v", err)
	}

	nir_encoded, err := json.Marshal(nirProg)
	if err != nil {
		return fmt.Errorf("json encoding failed: %v", err)
	}

	outputPath := cmd.String("output")
	if outputPath == "" {
		outputPath = strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".json"
	}

	if err := os.WriteFile(outputPath, nir_encoded, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	fmt.Printf("\\(^_^)/ successfully compiled %s to %s\n", filePath, outputPath)
	return nil
}
