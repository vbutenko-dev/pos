package files

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"example.org/pos/internal/application"
	"example.org/pos/internal/domain"
)

// Store хранит путь и исходную версию файла для одного сеанса.
type Store struct {
	path     string
	original []byte
	initial  []byte
}

type snapshot struct {
	Students []domain.Student   `json:"students"`
	Criteria []domain.Criterion `json:"criteria"`
	Grades   []gradeRow         `json:"grades"`
}

type gradeRow struct {
	StudentID string `json:"student_id"`
	Criterion string `json:"criterion"`
	Score     int    `json:"score"`
}

// Open загружает данные в пустую ведомость. Отсутствие файла означает первый запуск.
func Open(path string, book *application.Gradebook) (*Store, error) {
	data, err := readExisting(path)
	if err != nil {
		return nil, err
	}
	if data != nil {
		var state *snapshot
		if err := json.Unmarshal(data, &state); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if state == nil {
			return nil, fmt.Errorf("%s: ожидалась ведомость", path)
		}
		if err := state.restore(book); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
	}
	initial, err := encode(book)
	if err != nil {
		return nil, err
	}
	return &Store{path: path, original: data, initial: initial}, nil
}

func (s snapshot) restore(book *application.Gradebook) error {
	for _, student := range s.Students {
		if err := book.AddStudent(student.ID, student.Name); err != nil {
			return err
		}
	}
	for _, criterion := range s.Criteria {
		if err := book.AddCriterion(criterion); err != nil {
			return err
		}
	}
	for _, grade := range s.Grades {
		if err := book.SetScore(grade.StudentID, grade.Criterion, grade.Score); err != nil {
			return err
		}
	}
	return nil
}

func encode(book *application.Gradebook) ([]byte, error) {
	state := snapshot{Students: book.Students(), Criteria: book.Criteria(), Grades: []gradeRow{}}
	for _, student := range state.Students {
		_, grades, err := book.Results(student.ID)
		if err != nil {
			return nil, err
		}
		for _, e := range grades {
			state.Grades = append(state.Grades, gradeRow{StudentID: student.ID, Criterion: e.CriterionName(), Score: e.Score()})
		}
	}
	return json.MarshalIndent(state, "", "  ")
}

func readExisting(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return data, err
}

// Save сохраняет изменения, если файл по-прежнему соответствует открытой версии.
func (s *Store) Save(book *application.Gradebook) error {
	data, err := encode(book)
	if err != nil {
		return err
	}
	if bytes.Equal(s.initial, data) {
		return nil
	}
	err = replaceFile(s.path, func(w io.Writer) error {
		current, err := readExisting(s.path)
		if err != nil {
			return err
		}
		if (current == nil) != (s.original == nil) || !bytes.Equal(current, s.original) {
			return fmt.Errorf("файл %s изменился с момента открытия ведомости", s.path)
		}
		_, err = w.Write(data)
		return err
	})
	if err == nil {
		s.original, s.initial = data, data
	}
	return err
}
