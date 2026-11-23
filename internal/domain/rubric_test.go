package domain_test

import (
	"errors"
	"testing"

	"example.org/pos/internal/domain"
)

func TestEvaluation(t *testing.T) {
	t.Parallel()
	rubric, err := domain.NewRubric(domain.Criterion{Name: "Лабораторная", MaxScore: 10, Weight: 60})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name      string
		criterion string
		score     int
		wantErr   error
	}{
		{"ноль", "Лабораторная", 0, nil},
		{"максимум", "Лабораторная", 10, nil},
		{"отрицательный балл", "Лабораторная", -1, domain.ErrInvalidScore},
		{"выше максимума", "Лабораторная", 12, domain.ErrInvalidScore},
		{"неизвестный критерий", "Экзамен", 8, domain.ErrUnknownCriterion},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := rubric.NewEvaluation("masha", tc.criterion, tc.score)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("ошибка = %v, ожидали %v", err, tc.wantErr)
			}
			if tc.wantErr == nil && (e.Score() != tc.score || e.StudentID() != "masha" || e.CriterionName() != tc.criterion) {
				t.Fatalf("неверная оценка: %+v", e)
			}
		})
	}
	e, err := rubric.NewEvaluation("masha", "Лабораторная", 8)
	if err != nil {
		t.Fatal(err)
	}
	if err := rubric.CorrectScore(e, 12); !errors.Is(err, domain.ErrInvalidScore) || e.Score() != 8 {
		t.Fatalf("после попытки поставить 12: балл = %d, ошибка = %v", e.Score(), err)
	}
	if err := rubric.CorrectScore(e, 9); err != nil || e.Score() != 9 {
		t.Fatalf("после исправления на 9: балл = %d, ошибка = %v", e.Score(), err)
	}
}
