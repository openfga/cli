package main

import (
	"fmt"
	"strings"

	"github.com/openfga/cli/internal/storetest"
)

func main() {
	// Create a test result with some sample data
	testResult := storetest.TestResults{
		Results: []storetest.TestResult{
			{
				Name: "Test 1",
				CheckResults: []storetest.ModelTestCheckSingleResult{
					{
						Expected:   true,
						Got:        func() *bool { b := true; return &b }(),
						TestResult: true,
					},
				},
			},
		},
	}

	fullOutput := testResult.FriendlyDisplay()
	friendlyBody := testResult.FriendlyBody()

	fmt.Println("Full output:")
	fmt.Println(fullOutput)
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Friendly body:")
	fmt.Println(friendlyBody)

	// Check if the header is properly removed
	if strings.Contains(friendlyBody, "# Test Summary #") {
		fmt.Println("\nERROR: Header not properly removed!")
	} else {
		fmt.Println("\nSUCCESS: Header properly removed!")
	}
}
