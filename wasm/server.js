var express = require('express');
var app = express();
express.static.mime.types['wasm'] = 'application/wasm';
app.use(express.static(__dirname + '/'));
port=8080
app.listen(8080);
console.log("server running at :8080")
