package lib

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/monocle/devcaddy/devcaddy/process"
)

var CommandMap = map[string]string{
	".go": "go run",
	".js": "node",
	".rb": "ruby",
}

type PluginConfig struct {
	Name, Command, Args string
	PipeTo, Path        string
	Opts                interface{}
	LogOnly, NoOutput   bool
}

func (cfg *PluginConfig) Parse() {
	if cfg.Command == "" {
		path := cfg.Path
		ext := filepath.Ext(path)
		cfg.Command = CommandMap[ext]

		if cfg.Command == "" {
			log.Fatalln(cfg.Args, ERROR_PLUGIN_COMMAND_UNKNOWN)
		}
	}

	name := cfg.Name
	if name == "" {
		if cfg.Command != "" && cfg.Path == "" {
			cfg.Name = cfg.Command
			return
		}
		base := filepath.Base(cfg.Path)
		split := strings.Split(base, ".")
		if len(split) <= 1 {
			name = cfg.Command
		} else {
			name = split[0]
		}
		cfg.Name = name
	}
}

func (cfg *PluginConfig) InjectedArgs(f *File) []string {
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
	for {
		in := <-p.InC
		if in.Op == ERROR {
			p.OutC <- in
		}

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
}

func NewPlugin(cfg *PluginConfig, fn func(*File) *File) *Plugin {
	p := &Plugin{
		PluginConfig: *cfg,
		Transform:    fn,
		InC:          make(chan *File),
		OutC:         make(chan *File),
	}

	go p.listen()
	return p
}

func NewIdentityPlugin() *Plugin {
	return NewPlugin(&PluginConfig{}, func(f *File) *File {
		return f
	})
}

func NewCommandPlugin(cfg *PluginConfig) *Plugin {
	cfg.Parse()

	fn := func(f *File) *File {
		args := cfg.InjectedArgs(f)

		if f.IsDeleted() {
			return NewFileWithContent(f.Name, "", f.Op)
		}

		cmd := exec.Command(cfg.Command, args...)
		output, err := cmd.CombinedOutput()
		return NewFileFromCommand(f, output, err, cfg.Name)
	}

	return NewPlugin(cfg, fn)
}

func NewProcessPlugin(cfg *PluginConfig) *Plugin {
	cfg.Parse()

	pluginDef, err := ioutil.ReadFile(cfg.Path)
	if err != nil {
		log.Fatalln("Error reading plugin file", err)
	}

	opts := ""
	if cfg.Opts != nil {
		opts = cfg.Opts.(string)
	}

	p := process.NewProcess(cfg.Command, string(pluginDef), opts)

	fn := func(f *File) *File {
		p.In <- process.Marshal(f.Name, f.Content)
		out := <-p.Out

		output := &File{
			Name:       f.Name,
			Op:         f.Op,
			PluginName: cfg.Name,
		}

		err = json.Unmarshal([]byte(out.Content), output)

		if err != nil {
			output.Error = err
		}

		if out.Error != nil {
			output.Error = out.Error
			output.Op = ERROR
		}

		return output
	}

	return NewPlugin(cfg, fn)
}

func NewPlugins(pcs []*PluginConfig) *Plugins {
	ps := map[string]*Plugin{}

	for _, conf := range pcs {
		var p *Plugin
		if filepath.Ext(conf.Path) == ".js" {
			p = NewProcessPlugin(conf)
		} else {
			p = NewCommandPlugin(conf)
		}

		if ps[p.Name] != nil {
			log.Fatalln(ERROR_PLUGIN_DUPLICATE)
		}
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
		log.Fatalln(ERROR_PLUGIN_NOT_DEFINED, name)
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
