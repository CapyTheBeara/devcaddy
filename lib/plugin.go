package lib

import (
	"os/exec"
	"strings"
)

const FILE_PATH_SPLITTER = "__SERVER_FILE_PATH__="

type PluginConfig struct {
	Name, Command, Args, PipeTo string
	LogOnly, NoOutput           bool
}

type Plugin struct {
	PluginConfig
	Transform func(*File) *File
	InC       chan *File
	OutC      chan *File
}

func (p *Plugin) listen() {
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

func NewPlugin(cfg *PluginConfig, fn func(*File) *File) *Plugin {
	p := &Plugin{
		PluginConfig: *cfg,
		Transform:    fn,
		InC:          make(chan *File),
		OutC:         make(chan *File),
	}

	p.listen()
	return p
}

func NewCommandPlugin(cfg *PluginConfig) *Plugin {
	args := strings.Split(cfg.Args, " ")

	fn := func(f *File) *File {
		res := &File{
			Name:    f.Name,
			LogOnly: f.LogOnly,
			Op:      f.Op,
		}

		cmd := exec.Command(cfg.Command, append(args, f.Name, f.Content)...)

		b, err := cmd.CombinedOutput()
		if err != nil {
			res.Error = err
		}

		split := strings.Split(string(b), FILE_PATH_SPLITTER)

		if len(split) > 1 {
			res.Name = strings.TrimSpace(split[1])
		}

		res.Content = split[0]
		return res
	}

	return NewPlugin(cfg, fn)
}
