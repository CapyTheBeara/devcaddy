package process

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os/exec"
	"strings"
	"text/template"
)

func logFatal(msg string, err error) {
	if err != nil {
		log.Fatal(msg+" error:", err)
	}
}

func init() {
	log.SetFlags(log.Lshortfile)
}

// need to run process from same dir as node_packages/
// err = os.Chdir("../../caddytest")
// logFatal("chdir", err)
func nodeScript(delim, modStr string) string {
	return `
process.stdin.setEncoding('utf8');

var END = "` + delim + `";
var buf = [];
var modStr = "` + template.JSEscapeString(modStr) + `";
var m;

if (modStr !== "") {
    m = new module.constructor();
    m.paths = module.paths;
    m._compile(modStr, '__devcaddy_node_plugin__.js');
}

process.stdin.on('data', function(chunk) {
    chunk = chunk.trim();

    if(chunk.indexOf(END) > 0) {
        chunk = chunk.replace(END, '');
        buf.push(chunk);
        res = buf.join('\n');

        try {
            if(m) {
                console.log(m.exports.plugin(res));
            } else {
                eval(res);
            }
        } catch (e) {
            process.stderr.write(e.message+'\n');
        }
        buf = [];
    } else {
        buf.push(chunk);
    }
});`

}

type Result struct {
	Content string
	Error   error
}

func NewProcess(name, delim, module string) *Process {
	cmd := exec.Command("node", "-e", nodeScript(delim, module))

	in, err := cmd.StdinPipe()
	logFatal("stdin pipe", err)

	out, err := cmd.StdoutPipe()
	logFatal("stdout pipe", err)
	outBuf := bufio.NewReader(out)

	e, err := cmd.StderrPipe()
	logFatal("stderr pipe", err)
	errBuf := bufio.NewReader(e)

	err = cmd.Start()
	logFatal("command start", err)

	p := Process{
		Name:   name,
		Cmd:    cmd,
		In:     make(chan string),
		Out:    make(chan *Result),
		InPipe: in,
		OutBuf: outBuf,
		ErrBuf: errBuf,
	}

	go p.listenIn()
	return &p
}

type Process struct {
	Name   string
	Cmd    *exec.Cmd
	In     chan string
	Out    chan *Result
	InPipe io.WriteCloser
	OutBuf *bufio.Reader
	ErrBuf *bufio.Reader
}

func (p *Process) listenIn() {
	for {
		input := <-p.In
		_, err := p.InPipe.Write([]byte(input + "\n"))
		logFatal("InPipe write", err)

		go p.listenOutBuf()
		go p.listenErrBuf()
	}
}

func (p *Process) listenOutBuf() {
	str := p.readyFromBuf(p.OutBuf)
	if str != "" {
		res := &Result{Content: str}
		p.Out <- res
	}
}

func (p *Process) listenErrBuf() {
	str := p.readyFromBuf(p.ErrBuf)
	if str != "" {
		res := &Result{Error: errors.New(str)}
		p.Out <- res
	}
}

func (p *Process) readyFromBuf(buf *bufio.Reader) string {
	res := []string{}

	for {
		out, err := buf.ReadString('\n')
		if err != nil && err != io.EOF {
			logFatal("Output buffer read", err)
			return ""
		}

		remaining := buf.Buffered()

		if err == io.EOF && remaining == 0 {
			return ""
		}

		trim := strings.TrimSpace(out)
		if trim != "" {
			res = append(res, trim)
		}

		if remaining == 0 && len(res) > 0 {
			return strings.Join(res, "\n")
		}
	}
}
