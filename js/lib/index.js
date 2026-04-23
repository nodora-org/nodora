import "./wasm_exec.js";
import { loadWasmBinary } from "#wasm-loader";

let wasmInstance = null;
let initPromise = null;

async function init() {
  if (initPromise) return initPromise;
  initPromise = (async () => {
    let GoClass = globalThis.Go;
    if (!GoClass) {
      throw new Error("wasm_exec.js did not define Go class");
    }
    const go = new GoClass();
    const wasmBytes = await loadWasmBinary();

    const result = await WebAssembly.instantiate(wasmBytes, go.importObject);
    wasmInstance = result.instance;
    go.run(wasmInstance);

    return wasmInstance;
  })();

  try {
    return await initPromise;
  } catch (error) {
    initPromise = null;
    wasmInstance = null;
    throw new Error(`init() failed: ${error.message}`);
  }
}

class Evaluator {
  constructor(evaluatorId) {
    this.evaluatorId = evaluatorId;
  }

  getId() {
    return this.evaluatorId;
  }

  evaluate(ruleName, input = {}) {
    const result = globalThis.__nodoraEvaluate(
      this.evaluatorId,
      ruleName,
      input,
    );
    if (result.error) throw new Error(result.error);
    return result;
  }

  async evaluateAsync(ruleName, input = {}) {
    return this.evaluate(ruleName, input);
  }

  on(signalName, callback) {
    if (typeof callback !== "function") {
      throw new Error("callback must be a function");
    }

    globalThis.__nodoraOnSignal(this.evaluatorId, signalName, callback);
    return this;
  }

  destroy() {
    globalThis.__nodoraDestroy(this.evaluatorId);
  }
}

async function createEvaluator(programJSON) {
  if (!wasmInstance) {
    await init();
  }

  const result = globalThis.__nodoraCreateEvaluator(programJSON);
  if (result.error) {
    throw new Error(result.error);
  }

  return new Evaluator(result);
}

async function compile(src) {
  if (!wasmInstance) {
    await init();
  }
  const result = globalThis.__nodoraCompile(src);
  if (result.error) throw result.error;
  return result;
}

function assertType(x, type, err) {
  if (!x || typeof x !== type) {
    throw new Error(err);
  }
}

async function registerFunction({ name, namespace, args, returnType, fn }) {
  if (!wasmInstance) await init();

  assertType(name, "string", "name is required and must be a string");
  assertType(
    returnType,
    "string",
    "returnType is required and must be a string",
  );
  assertType(fn, "function", "fn is required and must be a function");

  args = (args || []).map((arg, i) => {
    assertType(arg, "object", `args[${i}] must be an object`);
    assertType(
      arg.name,
      "string",
      `args[${i}].name is required and must be a string`,
    );
    assertType(
      arg.type,
      "string",
      `args[${i}].type is required and must be a string`,
    );

    return {
      name: arg.name,
      type: arg.type,
      required: !!(arg.required ?? true),
    };
  });

  namespace = typeof namespace === "string" ? namespace : "";
  const spec = {
    name,
    namespace,
    args: args,
    returnType,
  };

  const result = globalThis.__nodoraRegisterFunction(spec, fn);
  if (result && result.error) throw new Error(result.error);
}

export { compile, createEvaluator, Evaluator, registerFunction };
