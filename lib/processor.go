package lib

import (
	"os/exec"
)

type Processor struct {
	Transformer func(*File) *File
	SendOutput  bool
	InC         chan *File
	OutC        chan *File
}

func (p *Processor) listen() {
	go func() {
		for {
			select {
			case in := <-p.InC:
				out := &File{Name: in.Name}

				if p.SendOutput {
					out = p.Transformer(in)
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
		Transformer: fn,
		SendOutput:  sendOutput,
		InC:         make(chan *File),
		OutC:        make(chan *File),
	}

	p.listen()

	return p
}

func NewCommandProcessor(name string, opts ...interface{}) *Processor {
	args := []string{}

	if opts != nil {
		args = opts[0].([]string)
	}

	fn := func(f *File) *File {
		res := &File{Name: f.Name}
		cmd := exec.Command(name, append(args, f.Name, f.Content)...)

		out, err := cmd.Output()
		if err != nil {
			res.Error = err
		}

		res.Content = string(out)
		return res
	}

	return NewProcessor(fn)
}
