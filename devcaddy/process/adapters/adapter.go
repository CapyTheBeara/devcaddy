package adapters

import (
	"os/exec"
)

const DelimEnd = "__DEVCADDY_END__"
const DelimJoin = "__DEVCADDY_JOIN__"

var Map = map[string]Adapter{
	"node": Node,
}

type Adapter struct {
	Name string
	Args []string
	Fn   func(interface{}, interface{}) string
}

func (a *Adapter) Cmd(arg1, arg2 interface{}) *exec.Cmd {
	return exec.Command(a.Name, append(a.Args, a.Fn(arg1, arg2))...)
}
