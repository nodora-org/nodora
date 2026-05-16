export async function loadWasmBinary() {
  const wasmUrl = new URL("./nodora.wasm.gz", import.meta.url);
  const response = await fetch(wasmUrl.href);
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
  }
  if (response.headers.get("Content-Encoding")?.includes("gzip")) {
    return response.arrayBuffer();
  }
  const ds = new DecompressionStream("gzip");
  const decompressed = response.body.pipeThrough(ds);
  return await new Response(decompressed).arrayBuffer();
}
