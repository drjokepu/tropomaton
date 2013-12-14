module.exports = function(grunt) {
	grunt.loadNpmTasks('grunt-contrib-jshint');
	grunt.loadNpmTasks('grunt-contrib-less');
	grunt.registerTask('default', ['jshint']);
	grunt.config.init({
		jshint: {
			all: ['*.js']	
		},
		less: {
			development: {
				options: {
					report: 'min'
				},
				files: {
					'content/style.css': 'views/style.less'
				}
			}
		}
	});
};