var connect = require('connect');
var http = require('http');
var serveStatic = require('serve-static');
var openPage = require('open');

var server = connect();
server.use(serveStatic('./'));
http.createServer(server).listen(4001);

openPage('http://localhost:4001/tests/');
