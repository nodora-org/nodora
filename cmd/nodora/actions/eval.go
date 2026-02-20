package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/urfave/cli/v3"
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/evaluator"
	"nodora.org/nodora/pkg/nir"
)

func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%d ns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.2f µs", float64(d.Nanoseconds())/1000)
	}
	if d < time.Second {
		return fmt.Sprintf("%.2f ms", float64(d.Nanoseconds())/1e6)
	}
	return fmt.Sprintf("%.2f s", d.Seconds())
}

func Eval(ctx context.Context, cmd *cli.Command) error {
	filePath := cmd.String("file")

	var programRaw []byte
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("error reading file: %v", err)
		}
		programRaw = data
	}

	var p nir.Program
	err := json.Unmarshal([]byte(programRaw), &p)
	if err != nil {
		return err
	}

	var input core.ValueMap

	inputFile := cmd.String("input-file")
	if inputFile != "" {
		var inputData []byte
		inputData, err = os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("error reading input file: %v", err)
		}
		if err := json.Unmarshal(inputData, &input); err != nil {
			return err
		}
	}

	if inputFile == "" && cmd.Bool("stdin") {
		var inputData []byte
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return fmt.Errorf("missing input data, pipe JSON via stdin")
		}
		inputData, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		if err := json.Unmarshal(inputData, &input); err != nil {
			return err
		}
	}

	var wg sync.WaitGroup
	evaluator := evaluator.NewEvaluator(&p)
	evaluator.Debug = cmd.Bool("debug")

	execFlags := cmd.StringSlice("exec")
	if len(execFlags) > 0 {
		evaluator.SetWaitGroup(&wg)
	}

	for _, execFlag := range execFlags {
		parts := strings.SplitN(execFlag, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid exec format: %s (expected signal_name=command)", execFlag)
		}
		signalName := parts[0]
		command := parts[1]

		evaluator.OnSignal(signalName, func(args []any) error {
			cmdStr := command
			for i, arg := range args {
				placeholder := fmt.Sprintf("{%d}", i+1)
				cmdStr = strings.ReplaceAll(cmdStr, placeholder, fmt.Sprintf("%v", arg))
			}

			var shellCmd *exec.Cmd
			if runtime.GOOS == "windows" {
				shellCmd = exec.Command("cmd", "/c", cmdStr)
			} else {
				shellCmd = exec.Command("/bin/sh", "-c", cmdStr)
			}

			shellCmd.Stdout = os.Stdout
			shellCmd.Stderr = os.Stderr
			return shellCmd.Run()
		})
	}

	ruleName := cmd.String("rule")
	if ruleName == "" {
		rules := evaluator.GetRuleNames()
		if len(rules) > 1 {
			return fmt.Errorf("multiple rules found in %s: please specify which one to run using -r <rule_name>", filePath)
		}
		ruleName = rules[0]
	}

	start := time.Now()
	result, err := evaluator.EvaluateRule(ruleName, input)
	elapsed := time.Since(start)

	if err != nil {
		return err
	} else {
		var encoded, err = json.Marshal(result)
		if err != nil {
			return fmt.Errorf("JSON encoding error: %v", err)
		}
		fmt.Printf("\\(^_^)/ evaluation completed in %s\n---\n", formatDuration(elapsed))
		fmt.Println(string(encoded))
	}

	wg.Wait()
	return nil
}
