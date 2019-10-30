const http = require('http');
const WebSocket = require('ws');

function heartbeat() {
    clearTimeout(this.pingTimeout);

    // Use `WebSocket#terminate()`, which immediately destroys the connection,
    // instead of `WebSocket#close()`, which waits for the close timer.
    // Delay should be equal to the interval at which your server
    // sends out pings plus a conservative assumption of the latency.
    this.pingTimeout = setTimeout(() => {
	this.terminate();
    }, 30000 + 1000);
}

// make a request
var options = {
    port: 4242,
    hostname: '127.0.0.1',
    headers: {
	'Connection': 'keep-alive, Upgrade',
	'Upgrade': 'websocket'
    }
};

var req = http.request(options);

const client = new WebSocket('ws://127.0.0.1:4242/foo')
//const client = new WebSocket('ws://127.0.0.1:9000/foo')

client.on('message', function incoming(data) {
    console.log(data);
    client.send("from client");
});

client.on('open', function open() {
    console.log('open');
    client.send("hello");
    //heartbeat();
});

client.on('error', function (x) {
    console.log(x);
});

client.on('ping', function ping() {
    console.log('ping');
    heartbeat();
});

client.on('close', function clear() {
    console.log("close")
    clearTimeout(this.pingTimeout);
    client.terminate();
});
