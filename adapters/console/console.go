// Package console реализует текстовое меню преподавателя.
package console

import (
	"bufio"
	"cmp"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"example.org/pos/adapters/files"
	"example.org/pos/application"
	"example.org/pos/domain"
)

// Run обслуживает меню до выхода пользователя или конца ввода.
func Run(g *application.Gradebook, in io.Reader, out io.Writer) error {
	c := console{book: g, in: bufio.NewScanner(in), out: out}
	for {
		fmt.Fprintln(out, "\n1 — добавить студента\n2 — добавить критерий\n3 — выставить оценку\n4 — посмотреть оценки студента\n5 — список студентов\n6 — список критериев\n7 — исправить оценку\n8 — отменить оценку\n9 — рассчитать итог по весам\n10 — экспортировать итоги в totals.csv\n0 — сохранить ведомость и выйти")
		cmd, err := c.line("Выбор")
		if err == nil && cmd == "0" {
			return nil
		}
		if err == nil {
			err = c.execute(cmd)
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if c.in.Err() != nil {
			return c.in.Err()
		}
		if err != nil {
			fmt.Fprintln(out, "Ошибка:", err)
		}
	}
}

type console struct {
	book *application.Gradebook
	in   *bufio.Scanner
	out  io.Writer
}

func (c *console) line(label string) (string, error) {
	fmt.Fprintf(c.out, "%s: ", label)
	if c.in.Scan() {
		return strings.TrimSpace(c.in.Text()), nil
	}
	return "", cmp.Or(c.in.Err(), io.EOF)
}

func (c *console) fields(labels ...string) ([]string, error) {
	values := make([]string, 0, len(labels))
	for _, label := range labels {
		s, err := c.line(label)
		if err != nil {
			return nil, err
		}
		values = append(values, s)
	}
	return values, nil
}

func number(label, value string) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s: введите целое число", label)
	}
	return n, nil
}

func (c *console) execute(cmd string) error {
	switch cmd {
	case "1":
		v, err := c.fields("ID студента", "Имя студента")
		if err != nil {
			return err
		}
		if err := c.book.AddStudent(v[0], v[1]); err != nil {
			return err
		}
		fmt.Fprintln(c.out, "Студент добавлен.")
	case "2":
		v, err := c.fields("Название критерия", "Максимальный балл", "Вес")
		if err != nil {
			return err
		}
		maxScore, err := number("Максимальный балл", v[1])
		if err != nil {
			return err
		}
		weight, err := number("Вес", v[2])
		if err != nil {
			return err
		}
		if err := c.book.AddCriterion(domain.Criterion{Name: v[0], MaxScore: maxScore, Weight: weight}); err != nil {
			return err
		}
		fmt.Fprintln(c.out, "Критерий добавлен.")
	case "3", "7":
		v, err := c.fields("ID студента", "Название критерия", "Балл")
		if err != nil {
			return err
		}
		score, err := number("Балл", v[2])
		if err != nil {
			return err
		}
		save := c.book.SetScore
		if cmd == "7" {
			save = c.book.CorrectScore
		}
		if err := save(v[0], v[1], score); err != nil {
			return err
		}
		fmt.Fprintln(c.out, "Оценка сохранена.")
	case "4":
		id, err := c.line("ID студента")
		if err != nil {
			return err
		}
		s, grades, err := c.book.Results(id)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.out, "Оценки студента %s (%s):\n", s.Name, s.ID)
		if len(grades) == 0 {
			fmt.Fprintln(c.out, "Оценок пока нет.")
		}
		for _, e := range grades {
			fmt.Fprintf(c.out, "%s — %d\n", e.CriterionName(), e.Score())
		}
	case "5":
		students := c.book.Students()
		if len(students) == 0 {
			fmt.Fprintln(c.out, "Студентов пока нет.")
		}
		for _, s := range students {
			fmt.Fprintf(c.out, "%s — %s\n", s.ID, s.Name)
		}
	case "6":
		criteria := c.book.Criteria()
		if len(criteria) == 0 {
			fmt.Fprintln(c.out, "Критериев пока нет.")
		}
		for _, criterion := range criteria {
			fmt.Fprintf(c.out, "%s — максимум %d, вес %d\n", criterion.Name, criterion.MaxScore, criterion.Weight)
		}
	case "8":
		v, err := c.fields("ID студента", "Название критерия")
		if err != nil {
			return err
		}
		if err := c.book.CancelEvaluation(v[0], v[1]); err != nil {
			return err
		}
		fmt.Fprintln(c.out, "Оценка отменена.")
	case "9":
		id, err := c.line("ID студента")
		if err != nil {
			return err
		}
		totals, err := c.book.Totals(id)
		if err != nil {
			return err
		}
		c.printPercent("Текущий результат по выставленным оценкам", totals.Current, "оценок пока нет")
		c.printPercent("Итог по всем критериям курса", totals.Course, "критериев пока нет")
	case "10":
		if err := files.ExportCSV("totals.csv", c.book); err != nil {
			return err
		}
		fmt.Fprintln(c.out, "Итоги записаны в totals.csv.")
	default:
		return fmt.Errorf("неизвестный пункт меню %q", cmd)
	}
	return nil
}

func (c *console) printPercent(label string, percent *float64, empty string) {
	if percent == nil {
		fmt.Fprintf(c.out, "%s: %s\n", label, empty)
		return
	}
	fmt.Fprintf(c.out, "%s: %.2f%%\n", label, *percent)
}
