package process

import (
	"bufio"
	"errors"
	"io"
	"log"
	"strings"

	"github.com/monocle/devcaddy/devcaddy/process/adapters"
)

func logFatal(msg string, err error) {
	if err != nil {
		log.Fatal(msg+" error:", err)
	}
}

func Marshal(args ...string) string {
	return strings.Join(args, adapters.DelimJoin)
}

type Output struct {
	Content string
	Error   error
}

func NewProcess(name, arg1, arg2 string) *Process {
	adapter := adapters.Map[name]
	cmd := adapter.Cmd(arg1, arg2)
	delim := adapters.DelimEnd

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
		Delim:  delim,
		In:     make(chan string),
		Out:    make(chan *Output),
		inPipe: in,
		outBuf: outBuf,
		errBuf: errBuf,
		res:    make(chan *Output),
	}

	go p.listenIn()
	return &p
}

type Process struct {
	Delim  string
	In     chan string
	Out    chan *Output
	inPipe io.WriteCloser
	outBuf *bufio.Reader
	errBuf *bufio.Reader
	res    chan *Output
}

func (p *Process) listenIn() {
	for {
		in := <-p.In

		_, err := p.inPipe.Write([]byte(in + p.Delim + "\n"))
		logFatal("inPipe write", err)

		go p.listenOutBuf()
		go p.listenErrBuf()

		// block to avoid weird node error if multiple inputs
		// come in before first input is processed
		out := <-p.res
		p.Out <- out
	}
}

func (p *Process) listenOutBuf() {
	str := p.readFromBuf(p.outBuf)
	if str != "" {
		res := &Output{Content: str}
		p.res <- res
	}
}

func (p *Process) listenErrBuf() {
	str := p.readFromBuf(p.errBuf)
	if str != "" {
		res := &Output{Error: errors.New(str)}
		p.res <- res
	}
}

func (p *Process) readFromBuf(buf *bufio.Reader) string {
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
