package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Declare a cliCommand struct
type cliCommand struct {
	name        string
	description string
	callback    func() error
}

// Declare the command registry
var commands map[string]cliCommand

// Pokedex main function logic
func main() {

	// Define the command registry
	commands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
	}

	// Create new scanner for user input
	scanner := bufio.NewScanner(os.Stdin)

	// Infinite for loop to execute for each user command
	for {
		// Print prompt
		fmt.Print("Pokedex > ")

		// Process user command, clean input, and return first word
		if !scanner.Scan() {
			// Scanner encountered an error or EOF
			break
		}
		input := scanner.Text()

		// Get the command (first word)
		args := cleanInput(input)
		if len(args) == 0 {
			//Skip empty input and re-prompt
			continue
		}
		commandName := args[0]

		// Look up the command in the registry
		command, exists := commands[commandName]

		if exists {
			// If command exists, execute its callback
			err := command.callback()
			if err != nil {
				fmt.Println(err)
			}
		} else {
			// If command doesn't exist
			fmt.Println("Unknown command")
		}
	}

	// Check if the scanner stopped due to an error
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
}

// Function to clean up input (separate words and convert to lower case)
func cleanInput(text string) []string {

	// Process and return the input directly without an intermediate variable
	return strings.Fields(strings.ToLower(text))
}

// Function to handle 'exit' command
func commandExit() error {
	// Print exit prompt and exit program
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil // Line wont execute as os.Exit will terminate the program
}

// Function to handle 'help' command
func commandHelp() error {
	// Print help prompt
	fmt.Println("Welcome to the Pokedex!")
	fmt.Print("Usage:\n\n")

	// Loop over command registry, sort alphabetically, print command name and description
	var commandList []string
	for name := range commands {
		commandList = append(commandList, name)
	}
	sort.Strings(commandList)

	for _, name := range commandList {
		cmd := commands[name]
		fmt.Printf("%s: %s\n", name, cmd.description)
	}
	return nil
}
