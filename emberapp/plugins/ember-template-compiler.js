var compiler = require('ember-template-compiler');

var path = process.argv[2],
    file = process.argv[3],
    output = compiler.precompile(file).toString(),
    template = "Ember.Handlebars.template(" + output + ");\n",
    es6 = "import Ember from 'ember';\nexport default " + template;

console.log(es6 + '__SERVER_FILE_PATH__=' + path.replace('.hbs', '.js'));
