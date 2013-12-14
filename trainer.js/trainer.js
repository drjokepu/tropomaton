var express = require('express');
var path = require('path');
var Q = require('q');

var app = express();
app.set('view engine', 'jade');
app.set('views', path.join(__dirname, 'views'));
app.use('/content', express.static(path.join(__dirname, 'content')));

app.get('/', function(req, res) {
	res.render('index');
});

app.listen(8080);