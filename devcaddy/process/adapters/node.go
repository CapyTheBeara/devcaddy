package adapters

import (
	"text/template"
)

var Node = Adapter{
	Name: "node",
	Args: []string{"-e"},
	Fn:   nodeScript,
}

func nodeScript(arg1, arg2 interface{}) string {
	return `
process.stdin.setEncoding('utf8');

var END = "\` + DelimEnd + `";
var JOIN = "` + DelimJoin + `";
var buf = [];
var modStr = "` + template.JSEscapeString(arg1.(string)) + `";
var settings = "` + template.JSEscapeString(arg2.(string)) + `"
var m, input, split, res;

if (modStr !== "") {
    m = new module.constructor();
    m.paths = module.paths;
    m._compile(modStr, '__devcaddy_node_plugin__.js');
}

process.stdin.on('data', function(chunk) {
    chunk = chunk.trim();

    if(chunk.indexOf(END) > 0) {
        var _ch = chunk.replace(END, '');
        buf.push(_ch);
        input = buf.join('\n');
        buf = [];

        try {
            if(m) {
                if (!settings) {
                    settings = "{}";
                }

                split = input.split(JOIN);
                input = { name: split[0], content: split[1] };
                res = m.exports.plugin(input, JSON.parse(settings));
                console.log(JSON.stringify(res));
            } else {
                eval(input);
            }
        } catch (e) {
            process.stderr.write(e.message + '\n' + e.stack + '\n');
        }
    } else {
        buf.push(chunk);
    }
});`
}
