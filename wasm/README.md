# go-snark wasm
*Warning: this is an ongoing experimentation*

## Wasm usage
To compile to wasm, inside the `wasm` directory, execute:
```
GOARCH=wasm GOOS=js go build -o go-snark.wasm go-snark-wasm-wrapper.go
```

Add the file `wasm_exec.js` in the directory:
```
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
```

Call the library from javascript:
```js
let r = generateProofs(
	JSON.stringify(circuit),
	JSON.stringify(setup),
	JSON.stringify(px),
	JSON.stringify(inputs),
);
```

Run the http server that allows to load the `.wasm` file:
```
node server.js
```
