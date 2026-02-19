import { expect, test, describe } from "bun:test";
import { readdirSync, readFileSync, existsSync } from "fs";
import { join, dirname, basename } from "path";
import { fileURLToPath } from "url";

const { compile, createEvaluator } = require("../lib/index.js");

const __dirname = dirname(fileURLToPath(import.meta.url));
const testRoot = join(__dirname, "../../tests/samples");

function* walkDir(dir) {
  const entries = readdirSync(dir, { withFileTypes: true });
  for (const entry of entries) {
    const fullPath = join(dir, entry.name);
    if (entry.isDirectory()) {
      yield* walkDir(fullPath);
    } else {
      yield fullPath;
    }
  }
}

function getTestFiles() {
  const tests = [];
  for (const filePath of walkDir(testRoot)) {
    if (filePath.endsWith(".rule")) {
      const testName = basename(filePath, ".rule");
      const testDir = dirname(filePath);
      const inputsPath = join(testDir, `${testName}.inputs.json`);

      if (!existsSync(inputsPath)) {
        console.log(
          `Skipping ${testName}: no inputs file found at ${inputsPath}`,
        );
        continue;
      }

      tests.push({
        testName,
        rulePath: filePath,
        inputsPath,
      });
    }
  }
  return tests;
}

const testFiles = getTestFiles();

for (const { testName, rulePath, inputsPath } of testFiles) {
  describe(testName, async () => {
    const ruleContent = readFileSync(rulePath, "utf8");
    const inputsContent = readFileSync(inputsPath, "utf8");
    const testSamples = JSON.parse(inputsContent);

    const program = await compile(ruleContent);
    const programObj = JSON.parse(program);
    const ruleNames = Object.keys(programObj.rules || {});

    expect(ruleNames.length).toBeGreaterThan(0);

    test.each(testSamples.map((tc, i) => [i, tc]))(
      "sample_%i",
      async (_, tc) => {
        const program = await compile(ruleContent);
        const evaluator = await createEvaluator(program);

        try {
          if (tc.err) {
            expect(() => evaluator.evaluate(ruleNames[0], tc.input)).toThrow();
            return;
          }

          const result = evaluator.evaluate(ruleNames[0], tc.input);

          if (tc.expected.outputs) {
            for (const [key, expectedVal] of Object.entries(
              tc.expected.outputs,
            )) {
              const actualVal = result.outputs[key];
              expect(actualVal).toBeDefined();

              if (
                typeof actualVal === "number" &&
                typeof expectedVal === "number"
              ) {
                expect(actualVal).toBeCloseTo(expectedVal, 10);
              } else {
                expect(actualVal).toEqual(expectedVal);
              }
            }

            for (const key of Object.keys(result.outputs)) {
              expect(tc.expected.outputs).toHaveProperty(key);
            }
          }

          const expectedSignals = tc.expected.emitted_signals || [];
          expect(result.emitted_signals).toEqual(expectedSignals);
        } finally {
          evaluator.destroy();
        }
      },
    );
  });
}
