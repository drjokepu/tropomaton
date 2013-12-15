var http = require('http');
var Q = require('q');
var querystring = require('querystring');

var db = require('./db.js');

var page = {
	stats: function(client) {
		result = {};
		return highLevelStats(client).then(function(hl) {
			result.totalPageCount = hl.totalPageCount ? parseInt(hl.totalPageCount, 10) : 0;
			result.trainedPageCount = hl.trainedPageCount ? parseInt(hl.trainedPageCount, 10) : 0;
			return trainingStats(client);
		})
		.then(function(ts) {
			result.classes = ts;
			return result;
		});
	},
	nextUntrained: function(client) {
		var commandText =
			"select id, url, title " +
			"from page " +
			"where id >= floor(random() * (select max(id) from page))::integer " +
			"and human_class is null " +
			"limit 1";
		
		return db.query(client, commandText, { }).then(function(result) {
			if (result.rows.length === 0) {
				return null;
			} else {
				return {
					id: result.rows[0].id,
					url: result.rows[0].url,
					title: result.rows[0].title
				};
			}
		})
		.then(makePopulateWithPageLinkStats(client));
	},
	postTraining: function(pageId, pageClass) {
		var deferred = Q.defer();
		var postData = querystring.stringify({
			pageId: pageId,
			class: pageClass
		});
		
		var request = http.request({
			host: 'localhost',
			port: '8877',
			method: 'POST',
			path: '/train',
			headers: {
				'Content-Type': 'application/x-www-form-urlencoded',
				'Content-Length': postData.length
			}
		}, function(res) {
			if (res.statusCode === 200) {
				deferred.resolve();
			} else {
				deferred.reject(new Error('Response: ' + res.statusCode));
			}
		});
		
		request.on('error', function(error) {
			deferred.reject(error);	
		});
		
		request.write(postData);
		request.end();
		
		return deferred.promise;
	}
};

function highLevelStats(client) {
	var commandText =
		"select " +
		"(select count(*) from page) as page_count, " +
		"(select count(*) from page where human_class is not null) as trained_count";
		
	return Q.fcall(function() {
		return db.query(client, commandText, { });
	})
	.then(function (result) {
		if (result.rows.length === 0) {
			return {
				totalPageCount: 0,
				trainedPageCount: 0
			};
		} else {
			return {
				totalPageCount: result.rows[0].page_count,
				trainedPageCount: result.rows[0].trained_count
			};
		}
	});
}

function trainingStats(client) {
	var commandText =
		"select human_class, count(*) as count from page where human_class is not null group by human_class";
		
	return Q.fcall(function() {
		return db.query(client, commandText, { });
	})
	.then(function (result) {
		stats = {};
		for (var rowIndex = 0; rowIndex < result.rows.length; rowIndex++) {
			stats[parseInt(result.rows[rowIndex].human_class, 10)] = {
				class: parseInt(result.rows[rowIndex].human_class, 10),
				count: parseInt(result.rows[rowIndex].count, 10)
			};
		}
		
		return stats;
	});
}

function makePopulateWithPageLinkStats(client) {
	return function(pageObj) {
		if (pageObj) {
			return pageLinkStats(client, pageObj.id).then(function(stats) {
				pageObj.linkStats = stats;
				return pageObj;
			});
		} else {
			return pageObj;
		}
	};
}

function pageLinkStats(client, pageId) {
	var commandText =
		"select (select count(*) from link where page_to = :page_id) inbound, (select count(*) from link where page_from = :page_id) outbound";
		
	return Q.fcall(function() {
		return db.query(client, commandText, { page_id: pageId });
	})
	.then(function (result) {
		if (result.rows.length === 0) {
			return {
				inbound: 0,
				outbound: 0
			};
		} else {
			return {
				inbound: result.rows[0].inbound,
				outbound: result.rows[0].outbound
			};
		}
	});
}

module.exports = page;