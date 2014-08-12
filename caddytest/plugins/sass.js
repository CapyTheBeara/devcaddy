var sass = require('node-sass');

var includePaths = process.argv[2].split(','),
    outFile = process.argv[3],
    path = process.argv[4],
    fileContent = process.argv[5];

sass.renderFile({
  data: fileContent,
  includePaths: includePaths,
  outFile: outFile,
  success: function(){},
  error: function(err) { process.stderr.write("Sass: " + err); }
});


