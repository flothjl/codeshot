package main

import (
	"fmt"
	"time"
)

type Task struct {
	Name      string
	Completed bool
	Due       time.Time
}

func NewTask(name string, due time.Time) *Task {
	return &Task{Name: name, Due: due}
}

func (t *Task) MarkDone() {
	t.Completed = true
}

func main() {
	tasks := []*Task{
		NewTask("Write README", time.Now().Add(24*time.Hour)),
		NewTask("Push to GitHub", time.Now().Add(48*time.Hour)),
	}

	for _, task := range tasks {
		if !task.Completed {
			fmt.Printf("TODO: %s (due %s)\n", task.Name, task.Due.Format("Jan 2 15:04"))
		}
	}

	tasks[0].MarkDone()
	fmt.Println("Marked first task as done.")
}
