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
    go.run(wasmModule)
    console.log('WASM loaded successfully');
}).catch(err => {
    console.error('WASM loading failed:', err);
});

const data = { "copiedData": { "args": { "bitworkc": "000000", "bitworkr": "6238", "mint_ticker": "sophon", "nonce": 9999999, "time": 1705918935 } }, "nonceStart": 0, "nonceEnd": 9999999, "workerOptions": { "electrumApi": { "baseUrl": "http://ep.atomicalneutron.com/proxy", "usePost": true, "isOpenFlag": false }, "satsbyte": 10, "address": "bc1pq9a5tkcc987mknndz5fgrsj9ateyu046v6majxnzwpkxwy2t87nqygunry", "opType": "dmt", "dmtOptions": { "mintAmount": 100000, "ticker": "sophon" } }, "fundingWIF": "L4cjYizvfRVpjLNjZDqTYuKD5fJugNoYkTkYFDjpw21UrL5E4JT1", "fundingUtxo": { "txid": "0000008674690288a63dd83588d3a765a45a02aa9b6954b7eba16daf58507006", "vout": 1, "status": { "confirmed": true, "block_height": 826688, "block_hash": "0000000000000000000169fe9dbf52a11829a4c730e38b3d81fe11f399f4ca3c", "block_time": 1705844320 }, "value": 17437, "index": 1 }, "fees": { "commitAndRevealFee": 2920, "commitAndRevealFeePlusOutputs": 102920, "revealFeePlusOutputs": 101810, "commitFeeOnly": 1110, "revealFeeOnly": 1810 }, "performBitworkForCommitTx": true, "workerBitworkInfoCommit": { "input_bitwork": "000000", "hex_bitwork": "000000", "prefix": "000000" }, "scriptP2TR": null, "hashLockP2TR": null, "workerBitworkInfoReveal": { "input_bitwork": "6238", "hex_bitwork": "6238", "prefix": "6238" }, "additionalOutputs": [{ "address": "bc1pq9a5tkcc987mknndz5fgrsj9ateyu046v6majxnzwpkxwy2t87nqygunry", "value": 100000 }] }

self.addEventListener('message', e => {
    const index = e.data;
    data.copiedData.args.time += index;
    const result = mine(JSON.stringify(data));
    console.log('result', result)
    self.postMessage(result);
});

