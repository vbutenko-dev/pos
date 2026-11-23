// Package application содержит сценарии работы с ведомостью курса.
package application

import (
	"errors"
	"fmt"
	"strings"

	"example.org/pos/internal/domain"
)

var (
	// ErrStudentExists означает, что ID уже занят.
	ErrStudentExists = errors.New("студент с таким ID уже есть")
	// ErrUnknownStudent означает, что студента с таким ID нет в ведомости.
	ErrUnknownStudent = errors.New("студент не найден")
	// ErrNoEvaluation означает, что у студента нет оценки по выбранному критерию.
	ErrNoEvaluation = errors.New("оценка не найдена")
	// ErrMissingName означает, что имя студента пусто.
	ErrMissingName = errors.New("не указано имя студента")
)

// Repository задаёт операции хранения студентов и оценок.
type Repository interface {
	AddStudent(domain.Student) error
	Student(id string) (domain.Student, bool)
	Students() []domain.Student
	Evaluation(studentID, criterion string) (domain.Evaluation, bool)
	DeleteEvaluation(studentID, criterion string) bool
	SaveEvaluation(domain.Evaluation)
	Evaluations(studentID string) []domain.Evaluation
}

// Gradebook выполняет действия преподавателя с ведомостью одного курса.
type Gradebook struct {
	repo   Repository
	rubric *domain.Rubric
}

// NewGradebook связывает сценарии с рубрикатором и хранилищем.
func NewGradebook(repo Repository, rubric *domain.Rubric) *Gradebook {
	return &Gradebook{repo: repo, rubric: rubric}
}

// AddStudent регистрирует студента.
func (g *Gradebook) AddStudent(id, name string) error {
	if strings.TrimSpace(id) == "" {
		return domain.ErrMissingStudent
	}
	if strings.TrimSpace(name) == "" {
		return ErrMissingName
	}
	return g.repo.AddStudent(domain.Student{ID: id, Name: name})
}

// AddCriterion регистрирует критерий в рубрикаторе курса.
func (g *Gradebook) AddCriterion(c domain.Criterion) error {
	return g.rubric.AddCriterion(c)
}

// SetScore выставляет балл по критерию зарегистрированному студенту.
func (g *Gradebook) SetScore(studentID, criterion string, score int) error {
	if _, exists := g.repo.Student(studentID); !exists {
		return fmt.Errorf("%w: %q", ErrUnknownStudent, studentID)
	}
	e, err := g.rubric.NewEvaluation(studentID, criterion, score)
	if err != nil {
		return err
	}
	g.repo.SaveEvaluation(*e)
	return nil
}

// Results возвращает студента и выставленные ему оценки.
func (g *Gradebook) Results(studentID string) (domain.Student, []domain.Evaluation, error) {
	s, exists := g.repo.Student(studentID)
	if !exists {
		return domain.Student{}, nil, fmt.Errorf("%w: %q", ErrUnknownStudent, studentID)
	}
	return s, g.repo.Evaluations(studentID), nil
}

// Students возвращает зарегистрированных студентов.
func (g *Gradebook) Students() []domain.Student {
	return g.repo.Students()
}

// Criteria возвращает критерии курса.
func (g *Gradebook) Criteria() []domain.Criterion {
	return g.rubric.Criteria()
}

// CorrectScore изменяет существующую оценку после проверки нового балла.
func (g *Gradebook) CorrectScore(studentID, criterion string, score int) error {
	e, exists := g.repo.Evaluation(studentID, criterion)
	if !exists {
		return ErrNoEvaluation
	}
	if err := g.rubric.CorrectScore(&e, score); err != nil {
		return err
	}
	g.repo.SaveEvaluation(e)
	return nil
}

// CancelEvaluation отменяет оценку по одному критерию.
func (g *Gradebook) CancelEvaluation(studentID, criterion string) error {
	if !g.repo.DeleteEvaluation(studentID, criterion) {
		return ErrNoEvaluation
	}
	return nil
}

// Totals рассчитывает текущий результат и итог курса для выбранного студента.
func (g *Gradebook) Totals(studentID string) (domain.Totals, error) {
	_, grades, err := g.Results(studentID)
	if err != nil {
		return domain.Totals{}, err
	}
	return domain.CalculateTotals(g.rubric, grades), nil
}
