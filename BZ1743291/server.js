const http = require('http');
const WebSocket = require('ws');
const url = require('url');

function serve(port) {
    // const server = http.createServer();

    const server = http.createServer(function (req, res) {
    	console.log(req.headers)
    	res.writeHead(200, {'Content-Type': 'text/plain'});
    	res.end("\nThere's no place like "+port+", headers: "+JSON.stringify(req.headers)+"\n");
    })

    const wss = new WebSocket.Server({ noServer: true });

    wss.on('connection', function connection(ws, req, client) {
	//const ip = req.headers['x-forwarded-for'].split(/\s*,\s*/)[0];
	console.log("connection from "+req.headers['x-forwarded-for']);

	ws.on('message', function incoming(msg) {
	    //console.log(`Received message ${msg} from user `+ws)
	    ws.send('you said: '+msg)
	});
	ws.on('close', function() {
	    console.log('gone')
	});
	ws.on('connection', function() {
	    console.log('open')
	});
    });

    server.on('upgrade', function upgrade(request, socket, head) {
	const pathname = url.parse(request.url).pathname;

	console.log(pathname);
	if (pathname === '/foo') {
	    wss.handleUpgrade(request, socket, head, function done(ws) {
		wss.emit('connection', ws, request);
	    });
	} else {
	    socket.destroy();
	}
    })

    server.listen(port)
    console.log("listening on port "+port)
}

serve(9000);
serve(9001);
serve(9002);
