package console_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"example.org/pos/internal/application"
	"example.org/pos/internal/console"
	"example.org/pos/internal/domain"
	"example.org/pos/internal/memory"
)

func TestRejectedScorePreservesGrade(t *testing.T) {
	t.Parallel()
	input := strings.Join([]string{
		"1", "masha", "Маша",
		"2", "Лабораторная работа", "10", "60",
		"3", "masha", "Лабораторная работа", "8",
		"3", "masha", "Лабораторная работа", "12",
		"4", "masha",
		"0",
	}, "\n")
	var out bytes.Buffer
	g := application.NewGradebook(memory.New(), new(domain.Rubric))
	if err := console.Run(g, strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Студент добавлен.", "Критерий добавлен.", "Оценка сохранена.",
		"Ошибка: балл вне диапазона", "Оценки студента Маша (masha):\nЛабораторная работа — 8\n",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("в выводе отсутствует %q", want)
		}
	}
}

func TestGradebookMenu(t *testing.T) {
	t.Parallel()
	input := strings.Join([]string{
		"5", "6",
		"1", "masha", "Маша", "1", "anna", "Анна",
		"2", "Экзамен", "20", "40", "2", "Лабораторная", "10", "60",
		"5", "6",
		"3", "masha", "Лабораторная", "8", "3", "masha", "Экзамен", "15",
		"3", "anna", "Лабораторная", "7",
		"7", "masha", "Лабораторная", "9", "7", "masha", "Лабораторная", "12",
		"4", "masha",
		"8", "masha", "Лабораторная", "8", "masha", "Лабораторная",
		"7", "masha", "Лабораторная", "3", "4", "masha", "4", "anna", "0",
	}, "\n")
	g := application.NewGradebook(memory.New(), new(domain.Rubric))
	var out bytes.Buffer
	if err := console.Run(g, strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	remaining := out.String()
	for _, want := range []string{
		"Студентов пока нет.", "Критериев пока нет.",
		"anna — Анна\nmasha — Маша\n",
		"Лабораторная — максимум 10, вес 60\nЭкзамен — максимум 20, вес 40\n",
		"Ошибка: балл вне диапазона",
		"Оценки студента Маша (masha):\nЛабораторная — 9\nЭкзамен — 15\n\n",
		"Оценка отменена.", "Ошибка: оценка не найдена", "Ошибка: оценка не найдена",
		"Оценки студента Маша (masha):\nЭкзамен — 15\n\n",
		"Оценки студента Анна (anna):\nЛабораторная — 7\n\n",
	} {
		_, rest, found := strings.Cut(remaining, want)
		if !found {
			t.Fatalf("в этом месте вывода ожидали %q", want)
		}
		remaining = rest
	}
}

func TestRunReturnsOutputError(t *testing.T) {
	t.Parallel()
	reader, writer := io.Pipe()
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	book := application.NewGradebook(memory.New(), new(domain.Rubric))
	err := console.Run(book, strings.NewReader("0\n"), writer)
	if !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("Run вернул %v, ожидали %v", err, io.ErrClosedPipe)
	}
}
