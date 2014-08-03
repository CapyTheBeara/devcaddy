package lib

import (
	"os/exec"
	"strings"
)

const FILE_PATH_SPLITTER = "__SERVER_OUTPUT_PATH__="

type ProcessorConfig struct {
	Name, Command, Args, PipeTo string
	LogOnly, NoOutput           bool
}

type Processor struct {
	ProcessorConfig
	Transform func(*File) *File
	InC       chan *File
	OutC      chan *File
}

func (p *Processor) listen() {
	go func() {
		for {
			select {
			case in := <-p.InC:
				var out *File

				if !p.NoOutput {
					in.LogOnly = p.LogOnly
					out = p.Transform(in)
				}

				p.OutC <- out
			}
		}
	}()
}

func NewProcessor(cfg *ProcessorConfig, fn func(*File) *File) *Processor {
	p := &Processor{
		ProcessorConfig: *cfg,
		Transform:       fn,
		InC:             make(chan *File),
		OutC:            make(chan *File),
	}

	p.listen()
	return p
}

func NewCommandProcessor(cfg *ProcessorConfig) *Processor {
	args := strings.Split(cfg.Args, " ")

	fn := func(f *File) *File {
		res := &File{
			Name:    f.Name,
			LogOnly: f.LogOnly,
		}

		cmd := exec.Command(cfg.Command, append(args, f.Name, f.Content)...)

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

	return NewProcessor(cfg, fn)
}
