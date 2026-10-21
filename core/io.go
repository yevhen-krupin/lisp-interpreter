package core

import (
	"log"
	"os"
	"path/filepath"
	"slices"
)

type Closeable interface {
	Close() error
}

type Scope struct {
	f *os.File
}

func (t Scope) Close() {
	log.SetOutput(os.Stdout)
	if err := t.f.Close(); err != nil {
		log.Fatal(err)
	}
}
func (t Scope) Name() string {
	return t.f.Name()
}

func ToDisk(dir string, file string) Scope {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if os.MkdirAll(dir, 0777) != nil {
				log.Fatal("error making a dir")
			}
		}
	}
	file = filepath.Join(dir, file)
	if _, err := os.Stat(dir + "/" + file); err == nil {
		os.Remove(file)
	}
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(f)
	return Scope{f}
}

func GetArgument(name string, alias string) (string, bool) {
	// 0 - executable name
	i := slices.Index(os.Args, name)
	if i < 1 {
		i = slices.Index(os.Args, alias)
	}
	// found, next entry should be the value
	if i >= 1 {
		if i > len(os.Args)-1 || os.Args[i+1][0] == '-' {
			log.Println("the argument", name, "or", alias, "was expected but not provided")
			return "", false
		}
		return os.Args[i+1], true
	}
	return "", false
}
