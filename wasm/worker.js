// worker.js
self.importScripts('wasm_exec.js'); // Go 运行时脚本

const go = new Go();
let wasmModule;

fetch('main.wasm').then(response =>
    response.arrayBuffer()
).then(bytes =>
    WebAssembly.instantiate(bytes, go.importObject)
).then(result => {
    wasmModule = result.instance;
    console.log('WASM loaded', wasmModule);
}).catch(err => {
    console.error('WASM loading failed:', err);
});

self.addEventListener('message', e => {
    const [a, b] = e.data;
    console.log(wasmModule)
    const result = wasmModule.exports.multiply(a, b);
    self.postMessage(result);
});

