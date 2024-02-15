package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	options []string
	pwd     string
)

// A message used to indicate that activity has occurred. In the real world (for
// example, chat) this would contain actual data.
type responseMsg string

func main() {
	form := huh.NewForm(
		huh.NewGroup(
			// Let the user select multiple options.
			huh.NewMultiSelect[string]().
				Title("Make options").
				Options(
					huh.NewOption("Kill instance", "kill"),
					huh.NewOption("Pull Changes", "pull").Selected(true),
					huh.NewOption("Make", "make").Selected(true),
					huh.NewOption("Run", "run").Selected(true),
				).
				// Limit(4). // there’s a 4 options limit!
				Value(&options),
		),
	)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	// Если нет в ОС переменной PGPASS, то запрашиваем пароль
	if os.Getenv("PGPASS") == "" {
		// Gather some final details.
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Enter db password...").
					Value(&pwd).
					Validate(func(str string) error {
						if str == "" {
							return errors.New("sorry, you must enter a password")
						}
						err := os.Setenv("PGPASS", str)
						if err != nil {
							return err
						}
						return nil
					}),
			),
		)

		err := form.Run()
		if err != nil {
			log.Fatal(err)
		}
	}

	p := tea.NewProgram(model{
		sub:         make(chan string),
		currCommand: options[0],
		spinner:     spinner.New(),
	})

	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
		os.Exit(1)
	}

}

// Simulate a process that sends events at an irregular interval in real time.
// In this case, we'll send events on the channel at a random interval between
// 100 to 1000 milliseconds. As a command, Bubble Tea will run this
// asynchronously.
func listenForActivity(sub chan string) tea.Cmd {
	return func() tea.Msg {
		for _, command := range options {
			var data bytes.Buffer
			switch {
			case command == "kill":
				sub <- command
				// Чтение содержимого файла RUNNING_PID
				pidData, err := os.ReadFile("RUNNING_PID")
				if err != nil {
					log.Fatalf("Cannot read RUNNING_PID file: %v", err)
				}
				pid := strings.TrimSpace(string(pidData))
				cmd := exec.Command("kill", "-2", pid)
				cmd.Stdout = &data
				err = cmd.Run()
				if err != nil {
					log.Fatal(command, " error:", err)
				}

			case command == "pull":
				sub <- command
				cmd := exec.Command("git", "pull", "origin", "main")
				cmd.Stdout = &data
				err := cmd.Run()
				if err != nil {
					log.Fatal(command, " error:", err)
				}

			case command == "make":
				sub <- command
				cmd := exec.Command("make", "linux")
				cmd.Stdout = &data
				err := cmd.Run()
				if err != nil {
					log.Fatal(command, " error:", err)
				}
			case command == "run":
				sub <- command
				cmd := exec.Command("make", "run")
				cmd.Stdout = &data
				err := cmd.Run()
				if err != nil {
					log.Fatal(command, " error:", err)
				}
			}
			log.Println(command, " output:", data.String())
		}

		return responseMsg("done")
	}
}

// A command that waits for the activity on a channel.
func waitForActivity(sub chan string) tea.Cmd {
	return func() tea.Msg {
		return responseMsg(<-sub)
	}
}

type model struct {
	sub         chan string // where we'll receive activity notifications
	currCommand string      // текущая команда
	spinner     spinner.Model
	quitting    bool
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		listenForActivity(m.sub), // generate activity
		waitForActivity(m.sub),   // wait for activity
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		if msg == tea.KeyCtrlC {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil
	case responseMsg:
		m.currCommand = string(msg.(responseMsg))
		if m.currCommand == "done" {
			m.quitting = true
			return m, tea.Quit
		}
		return m, waitForActivity(m.sub) // wait for next event
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m model) View() string {
	s := fmt.Sprintf("\n %s Processing: %s\n\n", m.spinner.View(), m.currCommand)
	if m.quitting {
		s += "\n"
	}
	return s
}
