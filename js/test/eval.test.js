import { expect, test, mock } from "bun:test";

test("compile-to-evaluate pipeline", async () => {
  const { compile, createEvaluator } = require("../lib/index.js");
  const testRuleset = await compile(
    `
    signal TestSignal(x, y)

    rule TestRule { 
        out result = 123 
        emit TestSignal(result, 456) 
    }
    `,
  );

  const evaluator = await createEvaluator(testRuleset);
  expect(evaluator.getId()).toBeGreaterThan(0);

  const mockCallback = mock((x, y) => {
    expect(x).toEqual(123);
    expect(y).toEqual(456);
  });

  evaluator.on("TestSignal", mockCallback);
  const result = await evaluator.evaluateAsync("TestRule", { x: 123 });

  expect(mockCallback).toHaveBeenCalled();
  expect(mockCallback).toHaveBeenCalledTimes(1);
  expect(mockCallback).toHaveBeenCalledWith(123, 456);

  expect(result).toEqual({
    outputs: {
      result: 123,
    },
    emitted_signals: [
      {
        name: "TestSignal",
        args: [123, 456],
      },
    ],
  });

  evaluator.destroy();
});

test("accepts a ruleset serialized to JSON", async () => {
  const { compile, createEvaluator } = require("../lib/index.js");
  const ruleset = await compile("rule TestRule { out result = 123 }");

  // the CLI writes the compiled ruleset to disk as a JSON string
  const evaluator = await createEvaluator(JSON.stringify(ruleset));
  const result = await evaluator.evaluateAsync("TestRule", {});

  expect(result.outputs.result).toEqual(123);
  evaluator.destroy();
});
