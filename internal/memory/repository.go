// Package memory реализует хранилище для одного сеанса консоли.
package memory

import (
	"maps"
	"slices"

	"example.org/pos/internal/application"
	"example.org/pos/internal/domain"
)

// Repository хранит студентов по ID, а их оценки — по названию критерия.
type Repository struct {
	students map[string]domain.Student
	grades   map[string]map[string]domain.Evaluation
}

// New создаёт пустое хранилище.
func New() *Repository {
	return &Repository{students: make(map[string]domain.Student), grades: make(map[string]map[string]domain.Evaluation)}
}

// AddStudent сохраняет нового студента с уникальным ID.
func (r *Repository) AddStudent(s domain.Student) error {
	if _, exists := r.students[s.ID]; exists {
		return application.ErrStudentExists
	}
	r.students[s.ID] = s
	return nil
}

// Student возвращает студента по ID.
func (r *Repository) Student(id string) (domain.Student, bool) {
	s, exists := r.students[id]
	return s, exists
}

// SaveEvaluation записывает текущий балл по критерию.
func (r *Repository) SaveEvaluation(e domain.Evaluation) {
	id := e.StudentID()
	if r.grades[id] == nil {
		r.grades[id] = make(map[string]domain.Evaluation)
	}
	r.grades[id][e.CriterionName()] = e
}

// Evaluations возвращает оценки в порядке названий критериев.
func (r *Repository) Evaluations(id string) []domain.Evaluation {
	grades := r.grades[id]
	result := make([]domain.Evaluation, 0, len(grades))
	for _, name := range slices.Sorted(maps.Keys(grades)) {
		result = append(result, grades[name])
	}
	return result
}

// Students возвращает студентов в порядке ID.
func (r *Repository) Students() []domain.Student {
	result := make([]domain.Student, 0, len(r.students))
	for _, id := range slices.Sorted(maps.Keys(r.students)) {
		result = append(result, r.students[id])
	}
	return result
}

// Evaluation находит оценку студента по названию критерия.
func (r *Repository) Evaluation(id, criterion string) (domain.Evaluation, bool) {
	e, exists := r.grades[id][criterion]
	return e, exists
}

// DeleteEvaluation удаляет оценку и сообщает, была ли она выставлена.
func (r *Repository) DeleteEvaluation(id, criterion string) bool {
	grades := r.grades[id]
	if _, exists := grades[criterion]; !exists {
		return false
	}
	delete(grades, criterion)
	return true
}
