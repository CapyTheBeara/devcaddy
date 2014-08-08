package lib

import (
	"path/filepath"
	"sort"
	"strings"
)

func NewStore(c *Config) *Store {
	store = &Store{
		Root:      c.Root,
		Files:     make(map[string]*File),
		Input:     make(chan *File),
		DidUpdate: make(chan *File),
	}

	for _, f := range c.Files {
		store.Files[f.Name] = f
	}

	return store
}

var store *Store

type Store struct {
	Root      string
	Files     map[string]*File
	Input     chan *File
	DidUpdate chan *File // TODO - rename to Output
}

func (s *Store) Put(name string, content string, args ...bool) {
	f := &File{Name: name, Content: content}
	s.Files[f.Name] = f

	update := true
	if args != nil {
		update = args[0]
	}

	if update {
		s.doUpdate(f)
	}
}

func (s *Store) PutFile(f *File) {
	s.Files[f.Name] = f
	s.doUpdate(f)
}

func (s *Store) Get(name string) string {
	f := s.GetFile(name)
	if f == nil {
		return ""
	}
	return f.Content
}

func (s *Store) GetFile(name string) *File {
	f := s.Files[name]
	if f == nil {
		return nil
	}

	if f.Type == "merge" {
		s.MergeStoreFiles(f)
	}

	return f
}

func (s *Store) Delete(name string) {
	delete(s.Files, name)
	s.doUpdate(&File{Name: name})
}

func (s *Store) DeleteFile(f *File) {
	delete(s.Files, f.Name)
	s.doUpdate(f)
}

func (s *Store) GetAll() (files []*File) {
	for _, n := range s.SortedFileNames() {
		files = append(files, s.GetFile(n))
	}
	return
}

func (s *Store) SortedFileNames() []string {
	names := []string{}
	for name, _ := range s.Files {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (s *Store) Listen() {
	go func() {
		for {
			f := <-s.Input
			if f == nil {
				continue
			}

			switch f.Op {
			case CREATE:
				s.PutFile(f)
			case WRITE:
				s.PutFile(f)
			case REMOVE:
				s.DeleteFile(f)
			case RENAME:
				s.DeleteFile(f)
			}
		}
	}()
}

func (s *Store) MergeStoreFiles(file *File) string {
	contents := []string{}
	dir := filepath.Join(s.Root, file.Dir)

	if len(file.Files) > 0 {
		for _, f := range file.Files {
			path := filepath.Join(dir, f)
			contents = append(contents, s.Get(path))
		}
	} else {
		for _, n := range s.SortedFileNames() {
			f := s.Files[n]
			if strings.Contains(f.Name, dir) && strings.HasSuffix(f.Name, "."+file.Ext) {
				contents = append(contents, f.Content)
			}
		}
	}

	file.Content = strings.Join(contents, "\n")
	return file.Content
}

func (s *Store) doUpdate(f *File) {
	go func() {
		s.DidUpdate <- f
	}()
}
