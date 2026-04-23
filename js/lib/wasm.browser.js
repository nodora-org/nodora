export async function loadWasmBinary() {
  const wasmUrl = new URL("./nodora.wasm", import.meta.url);
  const response = await fetch(wasmUrl.href);
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
  }
  return response.arrayBuffer();
}
