var path = require('path'),
    Compiler = require("es6-module-transpiler").Compiler,
    jsStringEscape = require('js-string-escape');


function wrapInEval (output, fileName) {
  return 'eval("' +
    jsStringEscape(output) +
    '//# sourceURL=' + jsStringEscape(fileName) +
    '");\n'
}

var filePath = process.argv[2],
    file = process.argv[3];

var projectName = path.basename(process.cwd()),
    module = filePath.match(/app\/(.+)\.[^\.]+$/)[1],
    newPath = path.join(projectName, module),
    output = (new Compiler(file, newPath)).toAMD();

process.stdout.write(wrapInEval(output, newPath + '.js'));
