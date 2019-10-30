const http = require('http');
const WebSocket = require('ws');

const url = 'ws://127.0.0.1/foo'
const connection = new WebSocket(url)

connection.onopen = () => {
  connection.send('hey') 
}

connection.onerror = (error) => {
  console.log(`WebSocket error: ${error}`)
}

connection.onmessage = (e) => {
  console.log(e.data)
  connection.send('hey') 
}
