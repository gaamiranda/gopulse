package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

// Action represents the user's choice
type Action int

const (
	ActionAccept Action = iota
	ActionEdit
	ActionCancel
)

// CommitResult holds the result of the commit confirmation
type CommitResult struct {
	Action  Action
	Message string
}

// PRResult holds the result of the PR confirmation
type PRResult struct {
	Action      Action
	Title       string
	Description string
}

// ConfirmCommit shows the commit message and asks for confirmation
func ConfirmCommit(message string) (*CommitResult, error) {
	fmt.Println("\nGenerated commit message:")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println(message)
	fmt.Println(strings.Repeat("-", 50))

	var choice string
	err := huh.NewSelect[string]().
		Title("What would you like to do?").
		Options(
			huh.NewOption("Accept", "accept"),
			huh.NewOption("Edit", "edit"),
			huh.NewOption("Cancel", "cancel"),
		).
		Value(&choice).
		Run()

	if err != nil {
		return nil, fmt.Errorf("prompt failed: %w", err)
	}

	result := &CommitResult{Message: message}

	switch choice {
	case "accept":
		result.Action = ActionAccept
	case "edit":
		result.Action = ActionEdit
		// Allow editing the message
		var editedMessage string
		err := huh.NewText().
			Title("Edit commit message").
			Value(&editedMessage).
			CharLimit(500).
			Run()
		if err != nil {
			return nil, fmt.Errorf("edit prompt failed: %w", err)
		}
		if editedMessage != "" {
			result.Message = strings.TrimSpace(editedMessage)
		}
	case "cancel":
		result.Action = ActionCancel
	}

	return result, nil
}

// ConfirmPR shows the PR details and asks for confirmation
func ConfirmPR(title, description string) (*PRResult, error) {
	fmt.Println("\nGenerated PR:")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Title: %s\n\n", title)
	fmt.Println("Description:")
	fmt.Println(description)
	fmt.Println(strings.Repeat("-", 50))

	var choice string
	err := huh.NewSelect[string]().
		Title("What would you like to do?").
		Options(
			huh.NewOption("Accept", "accept"),
			huh.NewOption("Edit", "edit"),
			huh.NewOption("Cancel", "cancel"),
		).
		Value(&choice).
		Run()

	if err != nil {
		return nil, fmt.Errorf("prompt failed: %w", err)
	}

	result := &PRResult{
		Title:       title,
		Description: description,
	}

	switch choice {
	case "accept":
		result.Action = ActionAccept
	case "edit":
		result.Action = ActionEdit
		// Allow editing title and description
		var newTitle, newDescription string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("PR Title").
					Value(&newTitle).
					Placeholder(title),
				huh.NewText().
					Title("PR Description").
					Value(&newDescription).
					CharLimit(2000),
			),
		)

		err := form.Run()
		if err != nil {
			return nil, fmt.Errorf("edit prompt failed: %w", err)
		}

		if newTitle != "" {
			result.Title = strings.TrimSpace(newTitle)
		}
		if newDescription != "" {
			result.Description = strings.TrimSpace(newDescription)
		}
	case "cancel":
		result.Action = ActionCancel
	}

	return result, nil
}

// ShowError displays an error message with formatting
func ShowError(err error) {
	fmt.Printf("\nError: %s\n", err.Error())
}

// ShowSuccess displays a success message
func ShowSuccess(message string) {
	fmt.Printf("\n%s\n", message)
}

// ShowInfo displays an informational message
func ShowInfo(message string) {
	fmt.Println(message)
}

// ShowSpinner displays a spinner with a message while an operation is in progress
// Returns a function to stop the spinner
func ShowSpinner(message string) func() {
	// For now, just print the message
	// In a future enhancement, we could use a proper spinner from bubbletea
	fmt.Printf("%s...\n", message)
	return func() {}
}
