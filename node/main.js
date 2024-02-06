const net = require('net');

var devices = new Array()

const client = new net.Socket();
client.connect(2281, '127.0.0.1', function() {
	console.log('Connected');
    // getDeviceID(data);
});

client.on('data', function(data) {
	console.log('Received ' + data.length + " bytes of data");
	// client.destroy(); // kill client after server's response
});

client.on('close', function() {
	console.log('Connection closed');
});
