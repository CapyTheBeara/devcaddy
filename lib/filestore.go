package lib

var store *Store

type Store struct {
	Root      string
	Files     []*File
	DidUpdate chan string
}

func (s *Store) Put(name string, content string, args ...bool) {
	f := &File{Name: name, Content: content}
	s.putFile(f)

	update := true
	if args != nil {
		update = args[0]
	}

	if update {
		s.doUpdate(name)
	}
}

func (s *Store) Get(name string) string {
	_, f := s.getFile(name)
	if f == nil {
		return ""
	}

	if f.Type == "merge" {
		return f.MergeStoreFiles(s)
	}

	return f.Content
}

func (s *Store) Delete(name string) {
	i, _ := s.getFile(name)
	if i == -1 {
		return
	}

	s.Files = append(s.Files[:i], s.Files[i+1:]...)
	s.doUpdate(name)
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
	i, file := s.getFile(f.Name)
	if i == -1 {
		s.Files = append(s.Files, f)
		return
	}

	file.Content = f.Content
}

func (s *Store) doUpdate(name string) {
	go func() {
		s.DidUpdate <- name
	}()
}

func NewStore(root string, config *Config) *Store {
	store = &Store{
		Root:      root,
		Files:     []*File{},
		DidUpdate: make(chan string),
	}

	for _, f := range config.Files {
		store.putFile(f)
	}

	return store
}
