const http = require('http')
const express = require('express')

const app = express();

app.get('/', (_, res) => {
  res.sendFile(__dirname + '/index.html')
});

const server = http.createServer(app);

server.listen(0, () => {
  console.log('server listen open: ' + `http://localhost:${server.address().port}`);
})
