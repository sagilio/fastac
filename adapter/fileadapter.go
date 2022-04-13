package adapter

import (
	"bufio"
	"errors"
	"os"

	"example.com/lessbin/model"
)

type FileAdapter struct {
	path string
}

func NewFileAdapter(path string) *FileAdapter {
	return &FileAdapter{path: path}
}

func (a *FileAdapter) LoadPolicy(model model.Model) error {
	file, err := os.Open(a.path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		LoadPolicyLine(scanner.Text(), model)
	}

	return scanner.Err()
}

func (a *FileAdapter) SavePolicy(model model.Model) error {
	return errors.New("not implemented")
}
