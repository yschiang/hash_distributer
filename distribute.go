package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"time"
)

// RouteRequest determines the backend group based on request_id and distribution percentages.
func RouteRequest(requestID string, distribution []int) (string, string) {
	// Step 1: Hash the request_id using MD5 and use the full hash value
	hash := md5.Sum([]byte(requestID))
	hashValue := binary.BigEndian.Uint64(hash[:8]) % 100 // Use the first 8 bytes for better entropy

	// Convert hash to a readable hex string
	hashString := fmt.Sprintf("%x", hash[:8])

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
func Test_Distribution(ids []string, distribution []int, iterations int, outputFilePath string) (map[string]int, bool) {
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return nil, false
	}
	defer outputFile.Close()

	overallGroupCounts := make(map[string]int)
	multipleGroups := false

	for _, id := range ids {
		groupCounts := make(map[string]int)
		var hashString string
		for i := 0; i < iterations; i++ {
			group, hash := RouteRequest(id, distribution)
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

	outputFilePath := fmt.Sprintf("testdata/Test1_SimpleDistribution_%d_%d.output", d1, d2)

	// Generate 1000 random IDs
	numIDs := 1000
	iterations := 10
	randomIDs := Test_GenerateRandomIDs(numIDs, "default")

	// Test distribution
	distribution := []int{d1, d2}
	fmt.Printf("\n===== Testing %d-%d Distribution =====\n", d1, d2)
	groupCounts, multipleGroups := Test_Distribution(randomIDs, distribution, iterations, outputFilePath)
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
		outputFilePath := fmt.Sprintf("testdata/Test2_HashWithVariations_%d_%d_%s.output", d1, d2, idType)
		fmt.Printf("\n===== Testing %d-%d Distribution with ID Type: %s =====\n", d1, d2, idType)
		ids := Test_GenerateRandomIDs(1000, idType) // Generate 1000 IDs of the current type
		groupCounts, multipleGroups := Test_Distribution(ids, distribution, iterations, outputFilePath)
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
				RouteRequest(id, distribution)
			}
		}
		duration := time.Since(start) // Calculate elapsed time
		totalComputations := 1000 * len(ids)
		averageTime := duration / time.Duration(totalComputations) // Calculate average time per computation

		fmt.Printf("Benchmark for ID Type %s: Total time = %v, Average time per computation = %v\n", idType, duration, averageTime)
	}
}

func main() {
	// Ensure testdata folder exists
	if _, err := os.Stat("testdata"); os.IsNotExist(err) {
		os.Mkdir("testdata", os.ModePerm)
	}

	// Run Test_SimpleDistribution with different distributions
	Test_SimpleDistribution(50, 50)
	Test_SimpleDistribution(70, 30)

	// Run Test_HashWithVariations to test various ID formats
	Test_HashWithVariations(50, 50)
	Test_HashWithVariations(70, 30)

	// Run Test_BenchmarkTiming to benchmark processing times
	Test_BenchmarkTiming(50, 50)
}
