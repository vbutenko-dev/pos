package files

import (
	"encoding/csv"
	"io"
	"strconv"

	"example.org/pos/application"
)

// ExportCSV выгружает по одной строке на студента. Пустое поле означает отсутствие итога.
func ExportCSV(path string, book *application.Gradebook) error {
	return replaceFile(path, func(out io.Writer) error {
		w := csv.NewWriter(out)
		if err := w.Write([]string{"student_id", "name", "current_percent", "course_percent"}); err != nil {
			return err
		}
		for _, student := range book.Students() {
			totals, err := book.Totals(student.ID)
			if err != nil {
				return err
			}
			row := []string{student.ID, student.Name, percent(totals.Current), percent(totals.Course)}
			if err := w.Write(row); err != nil {
				return err
			}
		}
		w.Flush()
		return w.Error()
	})
}

func percent(value *float64) string {
	if value == nil {
		return ""
	}
	return strconv.FormatFloat(*value, 'f', 2, 64)
}
