package pokecache

import (
	"sync"
	"time"
)

// Define cacheEntry struct to track each entry in cache
type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

// Define Cache struct to store each cacheEntry
type Cache struct {
	entries  map[string]cacheEntry
	mu       sync.Mutex
	interval time.Duration
}

// Function to create a new cache with specified time interval
func NewCache(interval time.Duration) *Cache {

	//Initialize new Cache struct and its entries map
	cache := Cache{
		entries:  make(map[string]cacheEntry),
		interval: interval,
	}

	// Initiate a goroutine for the reapLoop function
	go cache.reapLoop(interval)

	// Return address to Cache struct
	return &cache
}

// Helper function to add an entry to Cache
func (c *Cache) Add(key string, val []byte) {

	// Lock Cache mutex to ensure thread safety
	c.mu.Lock()
	// Defer Unlock for mutex
	defer c.mu.Unlock()

	// Create new cacheEntry with current time and provided value
	entry := cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}

	// Add to Cache map
	c.entries[key] = entry
}

// Helper function to get an entry from Cache
func (c *Cache) Get(key string) ([]byte, bool) {

	// Lock Cache mutex to ensure thread safety
	c.mu.Lock()
	// Defer Unlock for mutex
	defer c.mu.Unlock()

	// Check if the key exists in the map
	entry, exists := c.entries[key]

	// Return nil, false if not found
	if !exists {
		return nil, false
	}

	// Return entry value and true
	return entry.val, true
}

// Helper function to remove entries older than the specified interval
func (c *Cache) reapLoop(interval time.Duration) {

	// Create a ticker that triggers every interval
	ticker := time.NewTicker(interval)

	//Cleanup ticker when done
	defer ticker.Stop()

	// In a loop, wait for the ticker to trigger
	for range ticker.C {

		// Lock the mutex for thread safety
		c.mu.Lock()

		// Get current time for comparison
		now := time.Now()

		//Check each entry in the map to see if it's older than the interval
		for key, entry := range c.entries {

			//If the entry is older than the interval, remove it
			if now.Sub(entry.createdAt) > c.interval {
				delete(c.entries, key)
			}
		}

		//Unlock the mutex
		c.mu.Unlock()
	}
}
