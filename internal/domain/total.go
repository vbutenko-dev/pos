package domain

// Totals содержит проценты по оценённым работам и по всему курсу.
// nil означает, что для расчёта ещё нет критериев с подходящим весом.
type Totals struct {
	Current *float64
	Course  *float64
}

// CalculateTotals рассчитывает оба результата по текущим оценкам одного студента.
func CalculateTotals(rubric *Rubric, grades []Evaluation) Totals {
	scores := make(map[string]int, len(grades))
	for _, grade := range grades {
		scores[grade.CriterionName()] = grade.Score()
	}

	criteria := rubric.Criteria()
	return Totals{
		Current: weightedPercent(criteria, scores, gradedOnly),
		Course:  weightedPercent(criteria, scores, allCriteria),
	}
}

func gradedOnly(graded bool) bool { return graded }
func allCriteria(bool) bool       { return true }

func weightedPercent(criteria []Criterion, scores map[string]int, include func(bool) bool) *float64 {
	var earned, totalWeight float64
	for _, criterion := range criteria {
		score, graded := scores[criterion.Name]
		if !include(graded) {
			continue
		}
		weight := float64(criterion.Weight)
		earned += float64(score) / float64(criterion.MaxScore) * weight
		totalWeight += weight
	}
	if totalWeight == 0 {
		return nil
	}
	return new(100 * earned / totalWeight)
}
