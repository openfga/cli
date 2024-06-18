package generate

import (
	"errors"
	"fmt"
	"os"
	"path"
)

type GeneratedFile struct {
	Path             string
	Content          string
	OverrideIfExists bool
}

type GeneratedFiles struct {
	Files []GeneratedFile
}

func (f *GeneratedFiles) SaveAll(out string) error {
	for _, file := range f.Files {
		if err := file.Save(out); err != nil {
			return err
		}
	}
	return nil
}

func (f *GeneratedFile) Save(out string) error {
	if f.OverrideIfExists {
		if err := f.writeFile(out, f.Path, f.Content); err != nil {
			return err
		}
	} else {
		if err := f.writeFileIfNotExists(out, f.Path, f.Content); err != nil {
			return err
		}
	}
	return nil
}

func (f *GeneratedFile) writeFileIfNotExists(out, filePath, content string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get Working Directory %v", err)
	}

	file := path.Join(pwd, out, filePath)
	if _, err = os.Stat(file); errors.Is(err, os.ErrNotExist) {
		dir := path.Dir(file)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}

		f, err := os.Create(file)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = f.WriteString(content)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *GeneratedFile) writeFile(out, filePath, content string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get Working Directory %v", err)
	}

	fp := path.Join(pwd, out, filePath)
	dir := path.Dir(fp)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(fp, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(0644))
	if err != nil {
		return fmt.Errorf("failed to open or create file %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}
