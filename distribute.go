package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/cespare/xxhash/v2"
)

// Hasher defines the interface for different hashing strategies.
type Hasher interface {
	Hash(input string) uint64
}

// MD5Hasher implements the Hasher interface using MD5.
type MD5Hasher struct{}

func (h MD5Hasher) Hash(input string) uint64 {
	hash := md5.Sum([]byte(input))
	return binary.BigEndian.Uint64(hash[:8])
}

// XXHashHasher implements the Hasher interface using xxHash.
type XXHashHasher struct{}

func (h XXHashHasher) Hash(input string) uint64 {
	return xxhash.Sum64([]byte(input))
}

// DefaultHasher is the default hashing strategy (MD5).
var DefaultHasher Hasher = XXHashHasher{}

// RouteRequest determines the backend group based on request_id, distribution percentages, and hashing strategy.
func RouteRequest(requestID string, distribution []int, hasher ...Hasher) (string, string) {
	// Use the provided hasher or default to MD5Hasher
	chosenHasher := DefaultHasher
	if len(hasher) > 0 {
		chosenHasher = hasher[0]
	}

	// Step 1: Hash the request_id using the chosen hasher
	hashValue := chosenHasher.Hash(requestID) % 100 // Map to a percentage range

	// Convert hash to a readable hex string
	hashString := fmt.Sprintf("%x", hashValue)

	// Step 2: Determine the group based on the cumulative distribution
	cumulative := 0
	for index, percentage := range distribution {
		cumulative += percentage
		if int(hashValue) < cumulative {
			return fmt.Sprintf("Group %d", index+1), hashString
		}
	}

	return "Unknown Group", hashString // Fallback (should not occur with valid input)
}

// Test_Distribution tests a list of IDs and calculates the actual distribution.
func Test_Distribution(ids []string, distribution []int, iterations int, hasher Hasher, outputFile *os.File) (map[string]int, bool) {
	overallGroupCounts := make(map[string]int)
	multipleGroups := false

	for _, id := range ids {
		groupCounts := make(map[string]int)
		var hashString string
		for i := 0; i < iterations; i++ {
			group, hash := RouteRequest(id, distribution, hasher)
			hashString = hash
			groupCounts[group]++
			overallGroupCounts[group]++
		}

		// Check if any ID is distributed to more than one group
		if len(groupCounts) > 1 {
			multipleGroups = true
		}

		// Write detailed results for each ID to the output file
		summary := fmt.Sprintf("ID: %s [%s] -> ", id, hashString)
		for group, count := range groupCounts {
			summary += fmt.Sprintf("(%s: %d times) ", group, count)
		}
		outputFile.WriteString(summary + "\n")
	}

	return overallGroupCounts, multipleGroups
}

// Test_GenerateRandomIDs creates a list of random IDs based on type.
func Test_GenerateRandomIDs(count int, idType string) []string {
	ids := make([]string, count)
	for i := 0; i < count; i++ {
		switch idType {
		case "short_int":
			ids[i] = fmt.Sprintf("%d", rand.Intn(100))
		case "long_int":
			ids[i] = fmt.Sprintf("%d", rand.Int63n(1e12))
		case "tool":
			ids[i] = fmt.Sprintf("Tool-%d", rand.Intn(1000))
		case "string":
			ids[i] = fmt.Sprintf("ID-%s", randomString(8))
		default:
			ids[i] = fmt.Sprintf("ID-%d", rand.Intn(1000000)) // Default random ID
		}
	}
	return ids
}

// randomString generates a random string of given length.
func randomString(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	randomStr := make([]rune, length)
	for i := range randomStr {
		randomStr[i] = letters[rand.Intn(len(letters))]
	}
	return string(randomStr)
}

// Test_SimpleDistribution handles the logic for testing a simple distribution.
func Test_SimpleDistribution(d1, d2 int) {
	rand.Seed(time.Now().UnixNano()) // Seed random number generator

	// Open output file
	outputFile, err := os.Create(fmt.Sprintf("Test1_SimpleDistribution_%d_%d.output", d1, d2))
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	// Generate 1000 random IDs
	numIDs := 1000
	iterations := 10
	randomIDs := Test_GenerateRandomIDs(numIDs, "default")

	// Test distribution
	distribution := []int{d1, d2}
	fmt.Printf("\n===== Testing %d-%d Distribution =====\n", d1, d2)
	groupCounts, multipleGroups := Test_Distribution(randomIDs, distribution, iterations, DefaultHasher, outputFile)
	fmt.Printf("\nSummary of %d-%d Distribution:\n", d1, d2)
	if count, exists := groupCounts["Group 1"]; exists {
		fmt.Printf("Group 1: %d (%.2f%%)\n", count, float64(count)/(float64(numIDs)*float64(iterations))*100)
	}
	if count, exists := groupCounts["Group 2"]; exists {
		fmt.Printf("Group 2: %d (%.2f%%)\n", count, float64(count)/(float64(numIDs)*float64(iterations))*100)
	}
	fmt.Printf("Is there any ID distributed to more than 1 group? %t\n", multipleGroups)
}

// Test_HashWithVariations tests the hash function with various ID variations.
func Test_HashWithVariations(d1, d2 int) {
	rand.Seed(time.Now().UnixNano()) // Seed random number generator

	// ID types to test
	idTypes := []string{"short_int", "long_int", "tool", "string"}
	iterations := 10
	distribution := []int{d1, d2}

	for _, idType := range idTypes {
		outputFile, err := os.Create(fmt.Sprintf("Test2_HashWithVariations_%d_%d_%s.output", d1, d2, idType))
		if err != nil {
			fmt.Println("Error creating output file:", err)
			continue
		}
		defer outputFile.Close()

		fmt.Printf("\n===== Testing %d-%d Distribution with ID Type: %s =====\n", d1, d2, idType)
		ids := Test_GenerateRandomIDs(1000, idType) // Generate 1000 IDs of the current type
		groupCounts, multipleGroups := Test_Distribution(ids, distribution, iterations, DefaultHasher, outputFile)
		fmt.Printf("\nSummary for ID Type: %s\n", idType)
		if count, exists := groupCounts["Group 1"]; exists {
			fmt.Printf("Group 1: %d (%.2f%%)\n", count, float64(count)/(float64(len(ids)*iterations))*100)
		}
		if count, exists := groupCounts["Group 2"]; exists {
			fmt.Printf("Group 2: %d (%.2f%%)\n", count, float64(count)/(float64(len(ids)*iterations))*100)
		}
		fmt.Printf("Is there any ID distributed to more than 1 group? %t\n", multipleGroups)
	}
}

// Test_BenchmarkTiming benchmarks the time taken to process different types of IDs.
func Test_BenchmarkTiming(d1, d2 int) {
	rand.Seed(time.Now().UnixNano()) // Seed random number generator

	// ID types to test
	idTypes := []string{"short_int", "long_int", "tool", "string"}
	distribution := []int{d1, d2}

	for _, idType := range idTypes {
		ids := Test_GenerateRandomIDs(1000, idType) // Generate 1000 IDs of the current type

		start := time.Now()         // Start timing
		for i := 0; i < 1000; i++ { // Run 1000 iterations for benchmarking
			for _, id := range ids {
				RouteRequest(id, distribution, DefaultHasher)
			}
		}
		duration := time.Since(start) // Calculate elapsed time
		totalComputations := 1000 * len(ids)
		averageTime := duration / time.Duration(totalComputations) // Calculate average time per computation

		fmt.Printf("Benchmark for ID Type %s: Total time = %v, Average time per computation = %v\n", idType, duration, averageTime)
	}
}

func main() {
	// Run Test_SimpleDistribution with different distributions
	Test_SimpleDistribution(50, 50)
	Test_SimpleDistribution(70, 30)

	// Run Test_HashWithVariations to test various ID formats
	Test_HashWithVariations(50, 50)
	Test_HashWithVariations(70, 30)

	// Run Test_BenchmarkTiming to benchmark processing times
	Test_BenchmarkTiming(50, 50)
}
