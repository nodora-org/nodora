import { expect, test, mock } from "bun:test";

test("eval", async () => {
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
  expect(evaluator.getId()).toBe(1);

  const mockCallback = mock((x, y) => {
    expect(x).toEqual(123);
    expect(y).toEqual(456);
  });

  evaluator.on("TestSignal", mockCallback);
  const result = evaluator.evaluate("TestRule", { x: 123 });

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
