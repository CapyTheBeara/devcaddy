package lib

import (
	"os/exec"
	"strings"
)

const FILE_PATH_SPLITTER = "__SERVER_OUTPUT_PATH__="

type Processor struct {
	Name       string
	Transform  func(*File) *File
	SendOutput bool
	InC        chan *File
	OutC       chan *File
	PipeTo     string
}

func (p *Processor) listen() {
	go func() {
		for {
			select {
			case in := <-p.InC:
				out := &File{Name: in.Name}

				if p.SendOutput {
					out = p.Transform(in)
				}

				p.OutC <- out
			}
		}
	}()
}

func NewProcessor(fn func(*File) *File, opts ...interface{}) *Processor {
	sendOutput := true

	if opts != nil {
		if opts[0] != nil {
			sendOutput = opts[0].(bool)
		}
	}

	p := &Processor{
		Transform:  fn,
		SendOutput: sendOutput,
		InC:        make(chan *File),
		OutC:       make(chan *File),
	}

	p.listen()

	return p
}

func NewCommandProcessor(cmd string, opts ...interface{}) *Processor {
	args := []string{}
	sendOutput := true

	if opts != nil {
		args = opts[0].([]string)

		if len(opts) > 1 {
			sendOutput = opts[1].(bool)
		}
	}

	fn := func(f *File) *File {
		res := &File{Name: f.Name}
		cmd := exec.Command(cmd, append(args, f.Name, f.Content)...)

		b, err := cmd.Output()
		if err != nil {
			res.Error = err
		}

		split := strings.Split(string(b), FILE_PATH_SPLITTER)

		if len(split) > 1 {
			res.Name = split[1]
		}

		res.Content = split[0]
		return res
	}

	return NewProcessor(fn, sendOutput)
}
