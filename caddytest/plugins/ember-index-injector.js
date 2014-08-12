function cleanBaseURL(baseURL) {
  if (typeof baseURL !== 'string') { return; }
  if (baseURL[0] !== '/') { baseURL = '/' + baseURL; }
  if (baseURL.length > 1 && baseURL[baseURL.length - 1] !== '/') { baseURL = baseURL + '/'; }
  return baseURL;
};

function baseTag(){
  var baseURL      = cleanBaseURL(ENV.baseURL);
  var locationType = ENV.locationType;

  if (locationType === 'hash' || locationType === 'none') {
    return '';
  }

  if (baseURL) {
    return '<base href="' + baseURL + '" />';
  } else {
    return '';
  }
};

var file = process.argv[3];
var env = 'development';
var envFn = require('../config/environment.js');
var ENV = envFn(env);
var BASE_TAG = baseTag();

file = file.replace('{{BASE_TAG}}', BASE_TAG)
           .replace('{{ENV}}', JSON.stringify(ENV));

console.log(file);
