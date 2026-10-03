package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/caleb-parrot/quizez/internal/grok"
	"github.com/caleb-parrot/quizez/internal/quiz"
	"github.com/caleb-parrot/quizez/internal/tui"
)

func main() {
	args := os.Args[1:]
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Print(usage())
		return
	}
	client := grok.New()
	if len(args) >= 1 && args[0] == "-draw" {
		name := "Space"
		if len(args) >= 2 {
			name = strings.Join(args[1:], " ")
		}
		if err := drawOnce(client, name); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if len(args) != 0 {
		fmt.Fprintln(os.Stderr, usage())
		os.Exit(2)
	}
	if err := tui.Run(client); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func drawOnce(c *grok.Client, name string) error {
	cat, ok := quiz.ByName(name)
	if !ok {
		return fmt.Errorf("unknown category %q", name)
	}
	q, err := c.Draw(context.Background(), cat, nil)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n\n%s\n\n", cat.Name, q.Prompt)
	letters := []string{"A", "S", "D", "F"}
	for i, choice := range q.Choices {
		mark := " "
		if i == q.Answer {
			mark = "*"
		}
		fmt.Printf("%s %s  %s\n", mark, letters[i], choice)
	}
	return nil
}

func usage() string {
	return `quizez — one miss ends the run

  quizez                 category menu, then live questions
  quizez -draw Science   print one question and exit

Questions come from Grokipedia search at the moment they are asked.
Nothing is written to disk. A question that is built asks the Wayback Machine
to save the pages it used. The streak lasts until you quit.

TUI keys:
  1-9 space enter pick a category
  j/k             move
  a/s/d/f or 1-4  answer
  space enter     start, pick, or play again
  esc             back to categories
  q               quit
`
}
