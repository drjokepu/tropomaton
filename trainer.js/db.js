var inflect = require('i')();
var pg = require('pg');
var pgsubst = require('pgsubst');
var Q = require('q');

// pg://user:password@host:port/database
var connectionString = "pg:///tropomaton";

exports.run = function(fn) {
	var deferred = Q.defer();

	if (!fn) {
		process.nextTick(function() {
			var dbContext = null;
			Q.ninvoke(pg, 'connect', connectionString)
			.then(function(args) {
				dbContext = { client: args[0], done: args[1] };
				return beginTransaction(dbContext.client);
			})
			.then(function() {
				deferred.resolve(dbContext);
			}, function(err) {
				deferred.reject(err);
			}).done();
		});
	}
	else
	{
		var dbContext = null, result = null;
		exports.run().then(function(dbContext_arg) {
			dbContext = dbContext_arg;
			return fn(dbContext.client);
		})
		.then(function(result_arg) {
			result = result_arg;
			var promise = exports.closeConnection(dbContext);
			dbContext = null;
			return promise;
		})
		.then(function() {
			deferred.resolve(result);
		}, function(err) {
			deferred.reject(err);
		})
		.fin(function() {
			if (dbContext) {
				var promise = exports.cancelConnection(dbContext);
				dbContext = null;
				return promise;
			}
		}).done();
	}

	return deferred.promise;
};

exports.closeConnection = function(context) {
	var deferred = Q.defer();
	commitTransaction(context.client)
	.then(context.done)
	.then(function() {
		deferred.resolve();
	}, function(err) {
		deferred.reject(err);
	})
	.done();
	return deferred.promise;
};

exports.cancelConnection = function(context) {
	var deferred = Q.defer();
	rollbackTransaction(context.client)
	.then(context.done)
	.then(function() {
		deferred.resolve();
	}, function(err) {
		deferred.reject(err);
	})
	.done();
	return deferred.promise;
};

function beginTransaction(client) {
	var deferred = Q.defer();
	client.query('begin transaction', deferred.makeNodeResolver());
	return deferred.promise;
}

function commitTransaction(client) {
	var deferred = Q.defer();
	client.query('commit transaction', deferred.makeNodeResolver());
	return deferred.promise;
}

function rollbackTransaction(client) {
	var deferred = Q.defer();
	client.query('rollback transaction', deferred.makeNodeResolver());
	return deferred.promise;
}

exports.query = function(client, sql, params, print) {
	var deferred = Q.defer();
	process.nextTick(function() {
		var substSql = pgsubst(sql, params);
		if (print) console.log(substSql);
		Q.ninvoke(client, 'query', substSql)
		.then(function(result) {
			deferred.resolve(result);
		}, function(err) {
			deferred.reject(err);
		});
	});
	return deferred.promise;
};

exports.toCamelCase = function(obj) {
	if (!obj) return obj;
	if (obj.constructor.name === 'String') {
		return inflect.camelize(obj, false);
	} else {
		var copy = {};
		for (var key in obj) {
			copy[exports.toCamelCase(key)] = obj[key];
		}
		return copy;
	}
};

exports.parseArray = function(str) {
	if (str.constructor.name === 'String' && str[0] === '{' && str[str.length - 1] === '}') {
		return require('../../node_modules/pg/lib/types/arrayParser.js').create(str, null).parse(null);
	} else {
		return str;		
	}
};