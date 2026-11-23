package domain_test

import (
	"math"
	"testing"

	"example.org/pos/domain"
)

func TestWeightedTotals(t *testing.T) {
	t.Parallel()
	lab := domain.Criterion{Name: "Лабораторная", MaxScore: 10, Weight: 3}
	exam := domain.Criterion{Name: "Экзамен", MaxScore: 20, Weight: 2}
	course := []domain.Criterion{lab, exam}
	for _, tc := range []struct {
		name     string
		criteria []domain.Criterion
		scores   map[string]int
		current  *float64
		course   *float64
	}{
		{"пустой рубрикатор", nil, nil, nil, nil},
		{"нет оценок", course, nil, nil, new(0.0)},
		{"экзамен впереди", course, map[string]int{lab.Name: 8}, new(80.0), new(48.0)},
		{"за экзамен выставлен ноль", course, map[string]int{lab.Name: 8, exam.Name: 0}, new(48.0), new(48.0)},
		{"обе оценки выставлены", course, map[string]int{lab.Name: 8, exam.Name: 15}, new(78.0), new(78.0)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rubric, err := domain.NewRubric(tc.criteria...)
			if err != nil {
				t.Fatal(err)
			}
			var grades []domain.Evaluation
			for criterion, score := range tc.scores {
				e, err := rubric.NewEvaluation("masha", criterion, score)
				if err != nil {
					t.Fatal(err)
				}
				grades = append(grades, *e)
			}
			totals := domain.CalculateTotals(rubric, grades)
			checkPercent(t, "текущий", totals.Current, tc.current)
			checkPercent(t, "курс", totals.Course, tc.course)
		})
	}
}

func checkPercent(t *testing.T, label string, got, want *float64) {
	t.Helper()
	if got == nil || want == nil {
		if got != want {
			t.Fatalf("%s: есть результат = %t, ожидали %t", label, got != nil, want != nil)
		}
		return
	}
	if !(math.Abs(*got-*want) < 1e-9) {
		t.Errorf("%s = %v, ожидали %v", label, *got, *want)
	}
}
