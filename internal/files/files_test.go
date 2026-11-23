package files_test

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"example.org/pos/internal/application"
	"example.org/pos/internal/domain"
	"example.org/pos/internal/files"
	"example.org/pos/internal/memory"
)

func TestRoundTripAndExport(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "gradebook.json")
	book := application.NewGradebook(memory.New(), new(domain.Rubric))
	store, err := files.Open(path, book)
	if err != nil {
		t.Fatal(err)
	}
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(book.AddStudent("masha", "Маша, группа 1"))
	must(book.AddStudent("anna", "Анна"))
	must(book.AddCriterion(domain.Criterion{Name: "Лабораторная", MaxScore: 10, Weight: 60}))
	must(book.AddCriterion(domain.Criterion{Name: "Экзамен", MaxScore: 20, Weight: 40}))
	must(book.SetScore("masha", "Лабораторная", 9))
	must(book.CorrectScore("masha", "Лабораторная", 8))
	must(book.SetScore("masha", "Экзамен", 0))
	must(book.CancelEvaluation("masha", "Экзамен"))
	must(store.Save(book))

	restored := application.NewGradebook(memory.New(), new(domain.Rubric))
	_, err = files.Open(path, restored)
	must(err)
	report := filepath.Join(dir, "totals.csv")
	must(files.ExportCSV(report, restored))
	f, err := os.Open(report)
	must(err)
	defer func() { must(f.Close()) }()
	rows, err := csv.NewReader(f).ReadAll()
	must(err)
	want := [][]string{
		{"student_id", "name", "current_percent", "course_percent"},
		{"anna", "Анна", "", "0.00"},
		{"masha", "Маша, группа 1", "80.00", "48.00"},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("CSV = %v, ожидали %v", rows, want)
	}
}

func TestSavePreservesExternalChange(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "gradebook.json")
	book := application.NewGradebook(memory.New(), new(domain.Rubric))
	store, err := files.Open(path, book)
	if err != nil {
		t.Fatal(err)
	}
	if err := book.AddStudent("masha", "Маша"); err != nil {
		t.Fatal(err)
	}
	const other = `{"students":[{"ID":"anna","Name":"Анна"}],"criteria":[],"grades":[]}`
	if err := os.WriteFile(path, []byte(other), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(book); err == nil {
		t.Fatal("ожидалась ошибка изменения файла")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != other {
		t.Fatal("изменилась чужая версия ведомости")
	}
}
