import { readFile } from "fs/promises";
import { fileURLToPath } from "url";

export async function loadWasmBinary() {
  return readFile(fileURLToPath(new URL("./nodora.wasm", import.meta.url)));
}
