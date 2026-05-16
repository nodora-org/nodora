import { readFile } from "fs/promises";
import { fileURLToPath } from "url";
import { gunzip } from "zlib";
import { promisify } from "util";

const gunzipAsync = promisify(gunzip);

export async function loadWasmBinary() {
  const compressed = await readFile(
    fileURLToPath(new URL("./nodora.wasm.gz", import.meta.url))
  );
  return gunzipAsync(compressed);
}
