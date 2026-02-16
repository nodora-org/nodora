import { expect, test } from "bun:test";
const { compile } = require("../lib/index.js");

test("compile", async () => {
  const src = "rule Test{ out x = 1 }";
  const prog = await compile(src);
  expect(prog).toContain("Test");
});

test("compile with syntax error", async () => {
  const src = "rule Test{ a + 1 }";
  await expect(compile(src)).rejects.toThrow(/syntax error/);
});
