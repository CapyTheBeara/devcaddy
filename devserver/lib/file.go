package lib

import (
	"io/ioutil"
	"log"
)

const (
	CREATE FileOp = 1 << iota
	WRITE
	REMOVE
	RENAME
	CHMOD
	LOG
	ERROR
)

type FileOp uint32

func NewFile(name string, op FileOp) *File {
	f := File{Name: name, Op: op}
	if f.IsDeleted() {
		return &f
	}

	b, err := ioutil.ReadFile(f.Name)
	if err != nil {
		log.Fatalln(err)
	}

	f.Content = string(b)
	return &f
}

func NewFileWithContent(name, content string, op FileOp) *File {
	return &File{Name: name, Content: content, Op: op}
}

type File struct {
	Name        string
	Content     string
	Type        string
	Dir         string
	Ext         string
	Files       []string
	Error       error
	PluginNames []string `json:"plugins"`
	Op          FileOp
}

func (f *File) IsDeleted() bool {
	return f.Op == REMOVE || f.Op == RENAME
}
