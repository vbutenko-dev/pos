package main

import (
	"fmt"
	"os"

	"go.uber.org/dig"

	"example.org/pos/adapters/console"
	"example.org/pos/adapters/files"
	"example.org/pos/adapters/memory"
	"example.org/pos/application"
	"example.org/pos/domain"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	container := dig.New()
	if err := container.Provide(memory.New, dig.As(new(application.Repository))); err != nil {
		return err
	}
	if err := container.Provide(func() *domain.Rubric { return new(domain.Rubric) }); err != nil {
		return err
	}
	if err := container.Provide(application.NewGradebook); err != nil {
		return err
	}
	return container.Invoke(func(g *application.Gradebook) error {
		store, err := files.Open("gradebook.json", g)
		if err != nil {
			return err
		}
		if err := console.Run(g, os.Stdin, os.Stdout); err != nil {
			return err
		}
		return store.Save(g)
	})
}
