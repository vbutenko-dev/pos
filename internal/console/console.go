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

	"example.org/pos/internal/application"
	"example.org/pos/internal/domain"
	"example.org/pos/internal/files"
)

// Run обслуживает меню до выхода пользователя или конца ввода.
func Run(g *application.Gradebook, in io.Reader, out io.Writer) error {
	c := console{book: g, in: bufio.NewScanner(in), out: out}
	for {
		c.printf("\n1 — добавить студента\n2 — добавить критерий\n3 — выставить оценку\n4 — посмотреть оценки студента\n5 — список студентов\n6 — список критериев\n7 — исправить оценку\n8 — отменить оценку\n9 — рассчитать итог по весам\n10 — экспортировать итоги в totals.csv\n0 — сохранить ведомость и выйти\n")
		cmd, err := c.line("Выбор")
		if err == nil && cmd == "0" {
			return nil
		}
		if err == nil {
			err = c.execute(cmd)
		}
		if c.outputErr != nil {
			return c.outputErr
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if c.in.Err() != nil {
			return c.in.Err()
		}
		if err != nil {
			c.printf("Ошибка: %v\n", err)
		}
	}
}

type console struct {
	book      *application.Gradebook
	in        *bufio.Scanner
	out       io.Writer
	outputErr error
}

// printf сохраняет первую ошибку вывода, чтобы Run мог завершить сеанс.
func (c *console) printf(format string, args ...any) {
	if c.outputErr == nil {
		_, c.outputErr = fmt.Fprintf(c.out, format, args...)
	}
}

func (c *console) line(label string) (string, error) {
	c.printf("%s: ", label)
	if c.outputErr != nil {
		return "", c.outputErr
	}
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
		c.printf("Студент добавлен.\n")
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
		c.printf("Критерий добавлен.\n")
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
		c.printf("Оценка сохранена.\n")
	case "4":
		id, err := c.line("ID студента")
		if err != nil {
			return err
		}
		s, grades, err := c.book.Results(id)
		if err != nil {
			return err
		}
		c.printf("Оценки студента %s (%s):\n", s.Name, s.ID)
		if len(grades) == 0 {
			c.printf("Оценок пока нет.\n")
		}
		for _, e := range grades {
			c.printf("%s — %d\n", e.CriterionName(), e.Score())
		}
	case "5":
		students := c.book.Students()
		if len(students) == 0 {
			c.printf("Студентов пока нет.\n")
		}
		for _, s := range students {
			c.printf("%s — %s\n", s.ID, s.Name)
		}
	case "6":
		criteria := c.book.Criteria()
		if len(criteria) == 0 {
			c.printf("Критериев пока нет.\n")
		}
		for _, criterion := range criteria {
			c.printf("%s — максимум %d, вес %d\n", criterion.Name, criterion.MaxScore, criterion.Weight)
		}
	case "8":
		v, err := c.fields("ID студента", "Название критерия")
		if err != nil {
			return err
		}
		if err := c.book.CancelEvaluation(v[0], v[1]); err != nil {
			return err
		}
		c.printf("Оценка отменена.\n")
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
		c.printf("Итоги записаны в totals.csv.\n")
	default:
		return fmt.Errorf("неизвестный пункт меню %q", cmd)
	}
	return nil
}

func (c *console) printPercent(label string, percent *float64, empty string) {
	if percent == nil {
		c.printf("%s: %s\n", label, empty)
		return
	}
	c.printf("%s: %.2f%%\n", label, *percent)
}
