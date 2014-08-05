package lib

var store *Store

type Store struct {
	Root      string
	Files     map[string]*File
	DidUpdate chan string
}

func (s *Store) Put(name string, content string, args ...bool) {
	f := &File{Name: name, Content: content}
	s.Files[f.Name] = f

	update := true
	if args != nil {
		update = args[0]
	}

	if update {
		s.doUpdate(name)
	}
}

func (s *Store) Get(name string) string {
	f := s.Files[name]
	if f == nil {
		return ""
	}

	if f.Type == "merge" {
		return f.MergeStoreFiles(s)
	}

	return f.Content
}

func (s *Store) Delete(name string) {
	delete(s.Files, name)
	s.doUpdate(name)
}

func (s *Store) doUpdate(name string) {
	go func() {
		s.DidUpdate <- name
	}()
}

func NewStore(root string, config *Config) *Store {
	store = &Store{
		Root:      root,
		Files:     make(map[string]*File),
		DidUpdate: make(chan string),
	}

	for _, f := range config.Files {
		store.Files[f.Name] = f
	}

	return store
}
