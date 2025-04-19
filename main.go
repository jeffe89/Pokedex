package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"pokedex/internal/pokecache"
	"sort"
	"strings"
	"time"
)

// Define Config struct to track pagination URLs
type Config struct {
	Next     string
	Previous string
	Cache    *pokecache.Cache
}

// Define struct to parse JSON response
type LocationAreaResponse struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

// Define struct to hold Pokemon name and base experience
type Pokemon struct {
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`

	// Slice of stats, each with their own name and value
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`

	// Slice of types, each with their type name
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
}

// Define a cliCommand struct
type cliCommand struct {
	name        string
	description string
	callback    func(cfg *Config, args ...string) error
}

// Declare Pokedex caught Pokemon record
var caughtPokemon map[string]Pokemon

// Declare the command registry
var commands map[string]cliCommand

// Declare a global random generator
var rng *rand.Rand

// Pokedex main function logic
func main() {

	//Initialize rng seed for random logic
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	// Initialize the caught Pokemon map
	caughtPokemon = make(map[string]Pokemon)

	// Initialize the command registry
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
		"map": {
			name:        "map",
			description: "Display the names of 20 location areas in the Pokemon world",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Display the names of the previous 20 location areas in the Pokemon world",
			callback:    commandMapb,
		},
		"explore": {
			name:        "explore",
			description: "Explore a location area to find Pokemon",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Attempt to catch a Pokemon by name",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect Pokemon details",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Display all Pokemon found in Pokedex",
			callback:    commandPokedex,
		},
	}

	// Create new scanner for user input
	scanner := bufio.NewScanner(os.Stdin)

	// Create config with cache
	cfg := &Config{
		Cache: pokecache.NewCache(5 * time.Minute),
	}

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
			err := command.callback(cfg, args[1:]...)
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
func commandExit(cfg *Config, args ...string) error {

	//Warning to user indicating arguments are ignored
	if len(args) > 0 {
		fmt.Println("Warning: The exit command takes no arguments, ignoring extra input")
	}

	// Print exit prompt and exit program
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil // Line wont execute as os.Exit will terminate the program
}

// Function to handle 'help' command
func commandHelp(cfg *Config, args ...string) error {

	//Warning to user indicating arguments are ignored
	if len(args) > 0 {
		fmt.Println("Warning: The help command takes no arguments, ignoring extra input")
	}

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

// Function to handle 'map' command
func commandMap(cfg *Config, args ...string) error {

	//Warning to user indicating arguments are ignored
	if len(args) > 0 {
		fmt.Println("Warning: The map command takes no arguments, ignoring extra input")
	}

	// URL to request
	url := "https://pokeapi.co/api/v2/location-area"

	// If next URL from previous call available, use that instead
	if cfg.Next != "" {
		url = cfg.Next
	}

	// Check if data is in cache
	if cachedData, found := cfg.Cache.Get(url); found {
		fmt.Println("Using cached data...")

		// Parse the cached JSON response
		var locationResp LocationAreaResponse
		err := json.Unmarshal(cachedData, &locationResp)
		if err != nil {
			return err
		}

		// Update config with both next and previous URLs
		cfg.Next = locationResp.Next
		cfg.Previous = locationResp.Previous

		// Print out the location area names
		for _, location := range locationResp.Results {
			fmt.Println(location.Name)
		}

		return nil
	}

	// If not in cache, make HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Parse the JSON response
	var locationResp LocationAreaResponse
	err = json.Unmarshal(body, &locationResp)
	if err != nil {
		return err
	}

	// Update config with both next and previous URLs
	cfg.Next = locationResp.Next
	cfg.Previous = locationResp.Previous

	// Print out the location area names
	for _, location := range locationResp.Results {
		fmt.Println(location.Name)
	}

	return nil
}

// Function to handle 'mapb' command
func commandMapb(cfg *Config, args ...string) error {

	//Warning to user indicating arguments are ignored
	if len(args) > 0 {
		fmt.Println("Warning: The mapb command takes no arguments, ignoring extra input")
	}

	//Check if on the first page
	if cfg.Previous == "" {
		fmt.Println("you're on the first page.")
		return nil
	}

	// Set previous URL as the URL to request
	url := cfg.Previous

	// Check if data is in cache
	if cachedData, found := cfg.Cache.Get(url); found {
		fmt.Println("Using cached data...")

		// Parse the cached JSON response
		var locationResp LocationAreaResponse
		err := json.Unmarshal(cachedData, &locationResp)
		if err != nil {
			return err
		}

		// Update config with both next and previous URLs
		cfg.Next = locationResp.Next
		cfg.Previous = locationResp.Previous

		// Print out the location area names
		for _, location := range locationResp.Results {
			fmt.Println(location.Name)
		}

		return nil
	}

	// Make HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Parse the JSON response
	var locationResp LocationAreaResponse
	err = json.Unmarshal(body, &locationResp)
	if err != nil {
		return err
	}

	// Update config with both next and previous URLs
	cfg.Next = locationResp.Next
	cfg.Previous = locationResp.Previous

	// Print out the location area names
	for _, location := range locationResp.Results {
		fmt.Println(location.Name)
	}

	return nil
}

// Function to handle 'explore' command
func commandExplore(cfg *Config, location ...string) error {

	// Check if a location area name was provided
	if len(location) != 1 {
		return errors.New("you must provide a location area name")
	}

	// Use first argument as the location name
	locationName := location[0]

	// Warn user extra arguments will be ignored
	if len(location) > 1 {
		fmt.Println("Warning: Extra arguments provided will be ignored")
	}

	// Print exploring message
	fmt.Printf("Exploring %s...\n", locationName)

	// Check if cached
	if cachedPokemon, found := cfg.Cache.Get(locationName); found {
		// Unmarshal returned data
		var pokemonList []string
		if err := json.Unmarshal(cachedPokemon, &pokemonList); err != nil {
			return fmt.Errorf("failed to decode cached data for location %s: %w", locationName, err)
		}

		// If cached, display the Pokemon names
		fmt.Println("Found Pokemon:")
		for _, pokemon := range pokemonList {
			fmt.Printf(" - %s\n", pokemon)
		}
		return nil
	}

	// Fetch data from the PokeAPI
	fmt.Println("Fetching data from PokeAPI...")
	fetchedPokemon, err := fetchPokemonFromAPI(locationName)
	if err != nil {
		return fmt.Errorf("failed to explore %s: %w", locationName, err)
	}

	// Marshal the fetched data to JSON before caching
	fetchedPokemonJSON, err := json.Marshal(fetchedPokemon)
	if err != nil {
		return fmt.Errorf("failed to marshal Pokemon names: %w", err)
	}

	// Cache the result
	cfg.Cache.Add(locationName, fetchedPokemonJSON)

	// Display the fetched data
	fmt.Println("Found Pokemon:")
	for _, pokemon := range fetchedPokemon {
		fmt.Printf(" - %s\n", pokemon)
	}

	return nil
}

// Helper function to handle API request for location Pokemon
func fetchPokemonFromAPI(location string) ([]string, error) {

	// Define the PokeAPI endpoint for location-areas
	apiURL := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", location)

	// Make an HTTP GET request
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to PokeAPI: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-200 status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PokeAPI returned error code: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %w", err)
	}

	// Parse JSON into a map
	var responseData struct {
		PokemonEncounters []struct {
			Pokemon struct {
				Name string `json:"name"`
			} `json:"pokemon"`
		} `json:"pokemon_encounters"`
	}

	if err := json.Unmarshal(body, &responseData); err != nil {
		return nil, fmt.Errorf("failed to parse API response JSON: %w", err)
	}

	// Extract Pokemnon names from the API response
	var pokemonNames []string
	for _, encounter := range responseData.PokemonEncounters {
		pokemonNames = append(pokemonNames, encounter.Pokemon.Name)
	}

	// Check if no pokemon were found
	if len(pokemonNames) == 0 {
		return nil, errors.New("no pokemon found in this location area")
	}

	//Return the Pokemon names
	return pokemonNames, nil
}

// Function to handle 'catch' command
func commandCatch(cfg *Config, pokeName ...string) error {

	// Check if Pokemon name was provided in argument
	if len(pokeName) != 1 {
		return fmt.Errorf("please specify the name of the Pokemon to catch")
	}

	// Extract Pokemon name from parameter and print message to user
	pokemonName := pokeName[0]
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)

	// Define the PokeAPI endpoint for specified pokemon
	apiURL := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", pokemonName)

	// Make an HTTP GET request
	resp, err := http.Get(apiURL)
	if err != nil {
		return fmt.Errorf("failed to make request to PokeAPI: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-200 status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pokeAPI returned error code: %d", resp.StatusCode)
	}

	//Decore the JSON response body
	var pokemon Pokemon
	err = json.NewDecoder(resp.Body).Decode(&pokemon)
	if err != nil {
		return fmt.Errorf("failed to decode Pokemon data: %w", err)
	}

	// Determine the chance to catch the pokemon - Max = 100% / Min = 15% / Estimated Max Base Experiece = 350
	const (
		maxChance                  = 100
		minChance                  = 15
		estimatedMaxBaseExperience = 350
	)

	// Create a threshold by normalizing base experience ensuring catch rate between 100% to 15%
	threshold := maxChance - ((float64(pokemon.BaseExperience)) / (float64(estimatedMaxBaseExperience)) * (maxChance - minChance))

	// Given threshold - perform a random roll to determine if caught or not
	randChance := rng.Intn(100)
	if randChance > int(threshold) {
		fmt.Printf("%s escaped!\n", pokemonName)
		return nil
	} else {
		fmt.Printf("%s was caught!\n", pokemonName)
		caughtPokemon[pokemonName] = pokemon
		return nil
	}
}

// Function to handle 'inspect' command
func commandInspect(cfg *Config, pokeName ...string) error {

	// Check if Pokemon name was provided in argument
	if len(pokeName) != 1 {
		return fmt.Errorf("please specify the name of the Pokemon to inspect")
	}

	// Check if Pokemon has been caught - i.e. in the caughtPokemon map
	pokemonName := pokeName[0]
	pokemon, exists := caughtPokemon[pokemonName]
	if !exists {
		fmt.Printf("You have not caught the Pokemon %s.\n", pokemonName)
		return nil
	}

	// If caught - display Pokemon details
	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)

	fmt.Println("Stat:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("  - %s: %d\n", stat.Stat.Name, stat.BaseStat)
	}

	fmt.Println("Types:")
	for _, typeInfo := range pokemon.Types {
		fmt.Printf("  - %s\n", typeInfo.Type.Name)
	}

	return nil
}

// Function to handle `pokedex` command
func commandPokedex(cfg *Config, args ...string) error {

	// If caughtPokemon map is empty
	if len(caughtPokemon) == 0 {
		fmt.Println("You have not caught any Pokemon yet")
		return nil
	}

	// Print the names of all Pokemon in Pokedex
	fmt.Println("Your Pokedex:")
	for _, pokemon := range caughtPokemon {
		fmt.Printf(" - %s\n", pokemon.Name)
	}

	return nil
}
