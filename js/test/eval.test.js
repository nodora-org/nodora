import { expect, test, mock } from "bun:test";

test("compile-to-evaluate pipeline", async () => {
  const { compile, createEvaluator } = require("../lib/index.js");
  const testProgram = await compile(
    `
    signal TestSignal(x, y)

    rule TestRule { 
        out result = 123 
        emit TestSignal(result, 456) 
    }
    `,
  );

  const evaluator = await createEvaluator(testProgram);
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

test("accepts a program serialized to JSON", async () => {
  const { compile, createEvaluator } = require("../lib/index.js");
  const program = await compile("rule TestRule { out result = 123 }");

  // the CLI writes the compiled program to disk as a JSON string
  const evaluator = await createEvaluator(JSON.stringify(program));
  const result = await evaluator.evaluateAsync("TestRule", {});

  expect(result.outputs.result).toEqual(123);
  evaluator.destroy();
});
