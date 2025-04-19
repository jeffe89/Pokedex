# Pokedex 🧭
A command-line Pokédex written in Go, offering an interactive REPL (Read-Eval-Print Loop) that allows users to explore Pokémon data using the PokéAPI. It includes data caching and various commands to navigate, inspect, and catch Pokémon.

## Features
- REPL interface to issue commands like `map`, `explore`, `inspect`, and more  
- Caching mechanism to store API responses and reduce latency  
- Pagination support for browsing Pokémon locations  
- Ability to catch, inspect, and list Pokémon in your personal Pokédex  

## Getting Started

### Prerequisites
- Go 1.18 or later

### Installation
```bash
git clone https://github.com/jeffe89/Pokedex.git
cd Pokedex
go run main.go
```

## Usage
Start the application and use the following REPL commands:

- `help` – List all available commands  
- `map` – View current location and available areas  
- `mapb` – Move back one page in map listings  
- `explore <location>` – Explore a location for wild Pokémon  
- `catch <pokemon>` – Attempt to catch a Pokémon  
- `inspect <pokemon>` – View details of a caught Pokémon  
- `pokedex` – List caught Pokémon  
- `exit` – Quit the REPL  

## Project Structure
```
.
├── internal/
│   └── pokecache/
│       └── pokecache.go       # Provides timed caching for API responses
├── main.go                    # REPL and core application logic
├── repl_test.go               # Tests for REPL functionality
├── go.mod                     # Go module definition
├── LICENSE                    # MIT License
└── README.md                  # You're reading it!
```

## Key Functions

### main.go
- `startRepl()`  
  Initializes and runs the REPL interface, handling user input and command dispatch.

- `commandMap`  
  A map of command names to their corresponding handler functions and descriptions.

- `mapCommand(cfg *config)`  
  Fetches and displays the current paginated list of locations from the PokéAPI.

- `mapbCommand(cfg *config)`  
  Navigates backward one page in location listings.

- `exploreCommand(cfg *config, args []string)`  
  Retrieves Pokémon available at a specified location.

- `catchCommand(cfg *config, args []string)`  
  Attempts to catch a Pokémon and store it in the local Pokédex.

- `inspectCommand(cfg *config, args []string)`  
  Displays detailed information (height, weight, stats) of a caught Pokémon.

- `pokedexCommand(cfg *config)`  
  Lists all Pokémon the user has successfully caught.

### internal/pokecache/pokecache.go
- `NewCache(interval time.Duration)`  
  Creates a new timed cache for storing API data.

- `Add(key string, value []byte)`  
  Stores data in the cache under a specific key.

- `Get(key string)`  
  Retrieves data from the cache if it exists and is not expired.

## Running Tests
```bash
go test
```

## License
This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for more details.

## Author
Geoffrey Giordano

---
