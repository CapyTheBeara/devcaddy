package lib

import (
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func makeTestDir(t *testing.T, dir string, args ...interface{}) {
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		t.Fatal(err)
	}

	if args != nil {
		time.Sleep(time.Millisecond * time.Duration(args[0].(int)))
	}
}

func removeTestDir(t *testing.T, dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatal("Unable to remove dir", dir, err)
	}
}

func makeTestFile(t *testing.T, dir, name, content string, wait time.Duration) string {
	path := filepath.Join(dir, name)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(content)
	f.Sync()
	f.Close()

	time.Sleep(time.Millisecond * wait)
	return path
}

func updateTestFile(t *testing.T, name, content string) {
	f, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString(content)
	if err != nil {
		t.Fatal(err)
	}
	f.Sync()
	f.Close()
}
