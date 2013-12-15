var express = require('express');
var path = require('path');
var Q = require('q');

var db = require('./db.js');
var page = require('./page.js');

var app = express();
app.set('view engine', 'jade');
app.set('views', path.join(__dirname, 'views'));
app.use('/content', express.static(path.join(__dirname, 'content')));
app.use(express.bodyParser());

app.get('/', function(req, res) {
	var stats = null;
	db.run(function (client) {
		return page.stats(client).then(function(statsArg) {
			stats = statsArg;
		});
	})
	.then(function() {
		res.render('index', {
			model: {
				stats: stats
			}
		});	
	})
	.done();
});

app.get('/train', function(req, res) {
	var nextPage = null;
	db.run(function (client) {
		return page.nextUntrained(client).then(function(nextPageArg) {
			nextPage = nextPageArg;
		});
	})
	.then(function() {
		res.render('train', {
			model: {
				page: nextPage
			}
		});
	});
});

app.post('/train/submit', function(req, res) {
	page.postTraining(req.body.pageId, req.body.class).then(function() {
		res.redirect('/train');
	}).done();
});

app.listen(8080);