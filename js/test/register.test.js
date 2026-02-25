import { expect, test } from "bun:test";

test("register function", async () => {
  const {
    registerFunction,
    compile,
    createEvaluator,
  } = require("../lib/index.js");

  await registerFunction({
    namespace: "js",
    name: "add",
    args: [
      { name: "a", type: "number", required: true },
      { name: "b", type: "number", required: true },
    ],
    returnType: "number",
    fn: (a, b) => {
      return a + b;
    },
  });

  const testProgram = await compile(
    `
    rule TestRule { 
        out result = js::add(input.x, input.y) 
    }
    `,
  );

  const evaluator = await createEvaluator(testProgram);
  const result = await evaluator.evaluate("TestRule", { x: 1, y: 2 });

  expect(result).toEqual({
    outputs: {
      result: 3,
    },
    emitted_signals: [],
  });

  evaluator.destroy();
});

test("reject async function", async () => {
  const {
    registerFunction,
    compile,
    createEvaluator,
  } = require("../lib/index.js");

  await registerFunction({
    namespace: "js",
    name: "add_async",
    args: [
      { name: "a", type: "number", required: true },
      { name: "b", type: "number", required: true },
    ],
    returnType: "number",
    fn: async (a, b) => {
      return a + b;
    },
  });

  const testProgram = await compile(
    `
    rule TestRule { 
        out result = js::add_async(input.x, input.y) 
    }
    `,
  );

  const evaluator = await createEvaluator(testProgram);
  expect(evaluator.evaluate("TestRule", { x: 1, y: 2 })).rejects.toThrow(
    /async functions are not supported/,
  );

  evaluator.destroy();
});
