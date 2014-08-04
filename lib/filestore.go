package lib

var store *Store

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
	// TODO check if file is already present
	s.Files = append(s.Files, f)
}

func (s *Store) doUpdate() {
	go func() {
		s.DidUpdate <- true
	}()
}

func NewStore(root string, config *Config) *Store {
	store = &Store{
		Root:      root,
		Files:     []*File{},
		DidUpdate: make(chan bool),
	}

	for _, f := range config.Files {
		store.putFile(f)
	}

	return store
}
