// Package files сохраняет ведомость и выгружает итоги курса.
package files

import (
	"fmt"
	"io"
	"os"
)

// replaceFile публикует полностью записанный файл одним переименованием.
// Исключительное создание временного файла сериализует записи в один путь.
func replaceFile(path string, write func(io.Writer) error) (err error) {
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("не удалось начать запись %s: %w", path, err)
	}
	defer func() {
		if err != nil {
			os.Remove(tmp)
		}
	}()
	if err := write(f); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
