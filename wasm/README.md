# go-snark-study wasm
*Warning: this is an ongoing experimentation*

WASM wrappers for zkSNARK Pinocchio & Groth16 protocols.

## Wasm usage
To compile to wasm, inside the `wasm` directory, execute:
```
GOARCH=wasm GOOS=js go build -o go-snark.wasm go-snark-wasm-wrapper.go
```

Add the file `wasm_exec.js` in the directory:
```
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
```

To see the usage from javascript, check `index.js` file.

Run the http server that allows to load the `.wasm` file:
```
node server.js
```
