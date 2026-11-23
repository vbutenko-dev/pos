// Package domain описывает правила оценивания студентов.
package domain

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
)

var (
	// ErrInvalidCriterion означает, что у критерия отсутствует название, максимум или положительный вес.
	ErrInvalidCriterion = errors.New("нужно название критерия, положительный максимум и положительный вес")
	// ErrDuplicateCriterion означает, что рубрикатор уже содержит этот критерий.
	ErrDuplicateCriterion = errors.New("критерий уже существует")
	// ErrUnknownCriterion означает, что критерия нет в рубрикаторе.
	ErrUnknownCriterion = errors.New("критерия нет в рубрикаторе")
	// ErrInvalidScore означает, что балл выходит за допустимый диапазон.
	ErrInvalidScore = errors.New("балл вне диапазона")
	// ErrMissingStudent означает, что не указан идентификатор студента.
	ErrMissingStudent = errors.New("не указан студент")
)

// Criterion задаёт максимум и относительный вес работы.
type Criterion struct {
	Name     string
	MaxScore int
	Weight   int
}

// Rubric хранит критерии одного курса по названиям.
type Rubric struct {
	criteria map[string]Criterion
}

// NewRubric собирает рубрикатор из критериев с уникальными названиями.
func NewRubric(criteria ...Criterion) (*Rubric, error) {
	r := &Rubric{criteria: make(map[string]Criterion, len(criteria))}
	for _, c := range criteria {
		if err := r.AddCriterion(c); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// AddCriterion добавляет критерий с уникальным названием.
func (r *Rubric) AddCriterion(c Criterion) error {
	if strings.TrimSpace(c.Name) == "" || c.MaxScore <= 0 || c.Weight <= 0 {
		return ErrInvalidCriterion
	}
	if _, exists := r.criteria[c.Name]; exists {
		return fmt.Errorf("%w: %q", ErrDuplicateCriterion, c.Name)
	}
	if r.criteria == nil {
		r.criteria = make(map[string]Criterion)
	}
	r.criteria[c.Name] = c
	return nil
}

// Criteria возвращает критерии в порядке названий.
func (r *Rubric) Criteria() []Criterion {
	result := make([]Criterion, 0, len(r.criteria))
	for _, name := range slices.Sorted(maps.Keys(r.criteria)) {
		result = append(result, r.criteria[name])
	}
	return result
}

// Evaluation хранит балл студента за один критерий.
type Evaluation struct {
	studentID string
	criterion string
	score     int
}

// StudentID возвращает идентификатор студента.
func (e Evaluation) StudentID() string { return e.studentID }

// CriterionName возвращает название критерия.
func (e Evaluation) CriterionName() string { return e.criterion }

// Score возвращает выставленный балл.
func (e Evaluation) Score() int { return e.score }

// NewEvaluation выставляет первую оценку по критерию рубрикатора.
func (r *Rubric) NewEvaluation(studentID, criterion string, score int) (*Evaluation, error) {
	if strings.TrimSpace(studentID) == "" {
		return nil, ErrMissingStudent
	}
	e := &Evaluation{studentID: studentID, criterion: criterion}
	if err := r.CorrectScore(e, score); err != nil {
		return nil, err
	}
	return e, nil
}

// CorrectScore записывает новый балл после проверки критерия и диапазона.
func (r *Rubric) CorrectScore(e *Evaluation, score int) error {
	c, exists := r.criteria[e.criterion]
	if !exists {
		return fmt.Errorf("%w: %q", ErrUnknownCriterion, e.criterion)
	}
	if score < 0 || score > c.MaxScore {
		return fmt.Errorf("%w: получено %d, допустимо от 0 до %d", ErrInvalidScore, score, c.MaxScore)
	}
	e.score = score
	return nil
}
