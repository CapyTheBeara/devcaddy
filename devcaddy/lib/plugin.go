package lib

import (
	"log"
	"os/exec"
	"strings"
)

type PluginConfig struct {
	Name, Command, Args, PipeTo string
	LogOnly, NoOutput           bool
}

func (cfg *PluginConfig) ProcessCommandArgs(f *File) []string {
	argStr := cfg.Args
	if !strings.Contains(argStr, "{{fileName}}") && !strings.Contains(argStr, "{{fileContent}}") {
		argStr += " {{fileName}} {{fileContent}}"
	}

	// TODO - clean then use regexp.Split
	args := strings.Split(argStr, " ")
	for i, arg := range args {
		if strings.Contains(arg, "{{fileName}}") {
			args[i] = strings.Replace(arg, "{{fileName}}", f.Name, 1)
		} else {
			args[i] = strings.Replace(arg, "{{fileContent}}", f.Content, 1)
		}
	}
	return args
}

type Plugin struct {
	PluginConfig
	Transform func(*File) *File
	InC       chan *File
	OutC      chan *File
}

func (p *Plugin) SetOutC(c chan *File) {
	if p.PipeTo == "" {
		p.OutC = c
	}
}

func (p *Plugin) listen() {
	go func() {
		for {
			in := <-p.InC
			var out *File

			if !p.NoOutput {
				go func() {
					out = p.Transform(in)

					if p.LogOnly {
						out.Op = LOG
					}
					p.OutC <- out
				}()
			} else {
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

func NewIdentityPlugin() *Plugin {
	return NewPlugin(&PluginConfig{}, func(f *File) *File {
		return f
	})
}

func NewCommandPlugin(cfg *PluginConfig) *Plugin {
	fn := func(f *File) *File {
		args := cfg.ProcessCommandArgs(f)

		if f.IsDeleted() {
			return NewFileWithContent(f.Name, "", f.Op)
		}

		cmd := exec.Command(cfg.Command, args...)
		output, err := cmd.CombinedOutput()
		return NewFileFromCommand(f, output, err)
	}

	return NewPlugin(cfg, fn)
}

func NewPlugins(pcs []*PluginConfig) *Plugins {
	ps := map[string]*Plugin{}

	for _, conf := range pcs {
		p := NewCommandPlugin(conf)
		ps[p.Name] = p
	}

	plugins := Plugins{ps}

	for _, p := range ps {
		if p.PipeTo != "" {
			p.OutC = plugins.Get(p.PipeTo).InC
		}
	}

	return &plugins
}

type Plugins struct {
	content map[string]*Plugin
}

func (ps *Plugins) Get(name string) *Plugin {
	if name == "_identity_" {
		return NewIdentityPlugin()
	}

	p := ps.content[name]
	if p == nil {
		log.Fatalln("[error] Plugin was not defined:", name)
	}
	return p
}

func (ps *Plugins) Add(p *Plugin) {
	ps.content[p.Name] = p
}

func (ps *Plugins) Each(fn func(*Plugin)) int {
	i := 0
	for _, p := range ps.content {
		i++
		fn(p)
	}
	return i
}
