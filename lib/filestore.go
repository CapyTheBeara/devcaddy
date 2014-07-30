package lib

import (
	"encoding/json"
	"io"
	"path/filepath"
	"strings"
)

var store *Store

type Config struct {
	FileDefs []*File
}

type File struct {
	Name    string
	Content string
	Type    string
	Dir     string
	Ext     string
	Files   []string
}

func (file *File) mergeFiles() string {
	s := store
	contents := []string{}
	dir := filepath.Join(s.Root, file.Dir)

	if len(file.Files) > 0 {
		for _, f := range file.Files {
			path := filepath.Join(dir, f)
			contents = append(contents, s.Get(path))
		}
	} else {
		for _, f := range s.Files {
			if strings.Contains(f.Name, dir) && strings.HasSuffix(f.Name, "."+file.Ext) {
				contents = append(contents, f.Content)
			}
		}
	}

	return strings.Join(contents, "\n")
}

type Store struct {
	Root      string
	Files     []*File
	DidUpdate chan bool
}

func (s *Store) Put(name string, content string) {
	f := &File{Name: name, Content: content}
	s.putFile(f)
	s.doUpdate()
}

func (s *Store) Get(name string) string {
	_, f := s.getFile(name)
	if f == nil {
		return ""
	}

	if f.Type == "merge" {
		return f.mergeFiles()
	}

	return f.Content
}

func (s *Store) Delete(name string) {
	i, _ := s.getFile(name)
	if i == -1 {
		return
	}

	s.Files = append(s.Files[:i], s.Files[i+1:]...)
	s.doUpdate()
}

func (s *Store) getFile(name string) (i int, f *File) {
	for i, f = range s.Files {
		if name == f.Name {
			return i, f
		}
	}
	return -1, nil
}

func (s *Store) putFile(f *File) {
	s.Files = append(s.Files, f)
}

func (s *Store) doUpdate() {
	go func() {
		s.DidUpdate <- true
	}()
}

// func (s *Store) mapFiles(dir string, c *Config) {
// 	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
// 		if !info.IsDir() {
// 			if c.Ext != "" && !strings.HasSuffix(info.Name(), c.Ext) {
// 				return nil
// 			}

// 			f, err := ioutil.ReadFile(path)
// 			if err != nil {
// 				return err
// 			}

// 			content := string(f)

// 			if c.Processor != nil {
// 				path, content = c.Processor(path, content)
// 			}

// 			s.Put(path, content)
// 		}
// 		return nil
// 	})

// 	if err != nil {
// 		panic(err)
// 	}
// }

func NewStore(root string, cfg io.Reader) *Store {
	store = &Store{
		Root:      root,
		Files:     []*File{},
		DidUpdate: make(chan bool),
	}

	config := Config{}
	err := json.NewDecoder(cfg).Decode(&config)
	if err != nil {
		panic(err)
	}

	for _, f := range config.FileDefs {
		store.putFile(f)
	}

	return store
}
