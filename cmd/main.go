// main.go
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/earentir/inflation"

	cli "github.com/jawher/mow.cli"
)

func main() {
	app := cli.App("InflationCalculator", "A tool to calculate inflation-adjusted prices.")

	app.Spec = "[--inflation-list] [--cache]"

	// Define the --inflation-list flag
	inflationList := app.String(cli.StringOpt{
		Name:  "inflation-list",
		Desc:  "Path or URL to the inflation rate list JSON file",
		Value: "inflationratelist.json",
	})

	// Define the --cache flag
	cacheList := app.Bool(cli.BoolOpt{
		Name:  "cache",
		Desc:  "Cache the inflation list when downloaded from URL",
		Value: false,
	})

	// Command: yearInflation
	app.Command("year", "Get inflation rate for a specific year and country", func(cmd *cli.Cmd) {
		country := cmd.StringArg("COUNTRY", "", "Country name or code")
		dateStr := cmd.StringArg("DATE", "", "Date in YYYY or YYYY-MM format")

		cmd.Action = func() {
			if *country == "" || *dateStr == "" {
				fmt.Println("COUNTRY and DATE are required")
				cmd.PrintHelp()
				return
			}

			year, month, err := parseDate(*dateStr)
			if err != nil {
				log.Fatalf("Invalid DATE format: %v", err)
			}

			loader := &inflation.Loader{}
			err = loader.LoadData(*inflationList, *cacheList)
			if err != nil {
				log.Fatalf("Error loading data: %v", err)
			}

			rate, err := loader.Data.YearInflation(*country, year, month)
			if err != nil {
				log.Fatalf("Error fetching inflation rate: %v", err)
			}

			if month == 0 {
				fmt.Printf("Average inflation rate for %s in %d is %.2f%%\n", *country, year, rate)
			} else {
				fmt.Printf("Inflation rate for %s in %d-%02d is %.2f%%\n", *country, year, month, rate)
			}
		}
	})

	// Command: compareInflation
	app.Command("compare", "Compare inflation between two dates for a country", func(cmd *cli.Cmd) {
		country := cmd.StringArg("COUNTRY", "", "Country name or code")
		fromDateStr := cmd.StringArg("FROM_DATE", "", "From date in YYYY or YYYY-MM format")
		toDateStr := cmd.StringArg("TO_DATE", "", "To date in YYYY or YYYY-MM format")
		price := cmd.Float64Arg("PRICE", 0.0, "Original price") // Changed to Float64Arg

		cmd.Action = func() {
			if *country == "" || *fromDateStr == "" || *toDateStr == "" || *price == 0.0 {
				fmt.Println("COUNTRY, FROM_DATE, TO_DATE, and PRICE are required")
				cmd.PrintHelp()
				return
			}

			fromYear, fromMonth, err := parseDate(*fromDateStr)
			if err != nil {
				log.Fatalf("Invalid FROM_DATE format: %v", err)
			}

			toYear, toMonth, err := parseDate(*toDateStr)
			if err != nil {
				log.Fatalf("Invalid TO_DATE format: %v", err)
			}

			loader := &inflation.Loader{}
			err = loader.LoadData(*inflationList, *cacheList)
			if err != nil {
				log.Fatalf("Error loading data: %v", err)
			}

			newPrice, cumulativeRate, err := loader.Data.CompareInflation(*country, fromYear, fromMonth, toYear, toMonth, *price)
			if err != nil {
				log.Fatalf("Error comparing inflation: %v", err)
			}

			if fromMonth == 0 && toMonth == 0 {
				fmt.Printf("Price adjusted for inflation from %d to %d in %s: %.2f\nCumulative rate of inflation: %.2f%%\n", fromYear, toYear, *country, newPrice, cumulativeRate)
			} else if fromMonth != 0 && toMonth != 0 {
				fmt.Printf("Price adjusted for inflation from %d-%02d to %d-%02d in %s: %.2f\nCumulative rate of inflation: %.2f%%\n", fromYear, fromMonth, toYear, toMonth, *country, newPrice, cumulativeRate)
			} else if fromMonth != 0 && toMonth == 0 {
				fmt.Printf("Price adjusted for inflation from %d-%02d to %d in %s: %.2f\nCumulative rate of inflation: %.2f%%\n", fromYear, fromMonth, toYear, *country, newPrice, cumulativeRate)
			} else if fromMonth == 0 && toMonth != 0 {
				fmt.Printf("Price adjusted for inflation from %d to %d-%02d in %s: %.2f\nCumulative rate of inflation: %.2f%%\n", fromYear, toYear, toMonth, *country, newPrice, cumulativeRate)
			}
		}
	})

	// Command: compareWithBaseYear
	app.Command("compareWithBaseYear", "Compare inflation of a price relative to the country's Base Year", func(cmd *cli.Cmd) {
		country := cmd.StringArg("COUNTRY", "", "Country name or code")
		targetDateStr := cmd.StringArg("TARGET_DATE", "", "Target date in YYYY or YYYY-MM format")
		price := cmd.Float64Arg("PRICE", 0.0, "Original price") // Changed to Float64Arg

		cmd.Action = func() {
			if *country == "" || *targetDateStr == "" || *price == 0.0 {
				fmt.Println("COUNTRY, TARGET_DATE, and PRICE are required")
				cmd.PrintHelp()
				return
			}

			targetYear, targetMonth, err := parseDate(*targetDateStr)
			if err != nil {
				log.Fatalf("Invalid TARGET_DATE format: %v", err)
			}

			loader := &inflation.Loader{}
			err = loader.LoadData(*inflationList, *cacheList)
			if err != nil {
				log.Fatalf("Error loading data: %v", err)
			}

			newPrice, err := loader.Data.CompareInflationWithBaseYear(*country, targetYear, targetMonth, *price)
			if err != nil {
				log.Fatalf("Error comparing inflation with Base Year: %v", err)
			}

			countryData, err := loader.Data.GetCountry(*country)
			if err != nil {
				log.Fatalf("Error retrieving country data: %v", err)
			}

			if targetMonth == 0 {
				fmt.Printf("Price adjusted for inflation relative to Base Year (%d) to %d in %s: %.2f\n",
					countryData.BaseYear,
					targetYear,
					*country,
					newPrice)
			} else {
				fmt.Printf("Price adjusted for inflation relative to Base Year (%d) to %d-%02d in %s: %.2f\n",
					countryData.BaseYear,
					targetYear,
					targetMonth,
					*country,
					newPrice)
			}
		}
	})

	// Command: import
	app.Command("import", "Import inflation rates from a CSV file into a JSON file for a specific country", func(cmd *cli.Cmd) {
		country := cmd.StringArg("COUNTRY", "", "Country name or code")
		csvFile := cmd.StringArg("CSV_FILE", "", "Path to the CSV file with date,value")
		jsonFile := cmd.StringArg("JSON_FILE", "", "Path to the inflation JSON file to update")
		baseYear := cmd.Int(cli.IntOpt{
			Name:  "base-year",
			Desc:  "HICP Base Year for the country",
			Value: 2015, // Default Base Year
		})

		cmd.Action = func() {
			if *country == "" || *csvFile == "" || *jsonFile == "" {
				fmt.Println("COUNTRY, CSV_FILE, and JSON_FILE are required")
				cmd.PrintHelp()
				return
			}

			// Load existing JSON data
			loader := &inflation.Loader{}
			err := loader.LoadData(*jsonFile, false) // Not caching when loading
			if err != nil {
				log.Fatalf("Error loading JSON data: %v", err)
			}

			// Find the country; if not found, create a new one
			c, err := loader.Data.GetCountry(*country)
			if err != nil {
				// Country not found; create a new one
				fmt.Printf("Country '%s' not found. Creating a new country entry.\n", *country)
				newCountry := inflation.Country{
					Name:      *country, // Assuming country name is same as code; modify as needed
					Aliases:   []string{},
					Code:      *country,
					BaseYear:  *baseYear, // Set BaseYear
					Inflation: make(map[string]map[string]float64),
				}
				loader.Data.Countries = append(loader.Data.Countries, newCountry)
				// Retrieve the newly added country
				c, err = loader.Data.GetCountry(*country)
				if err != nil {
					log.Fatalf("Error creating new country: %v", err)
				}
			} else {
				if c.BaseYear == 0 {
					c.BaseYear = *baseYear // Set BaseYear if not already set
				}
			}

			// Read CSV
			file, err := os.Open(*csvFile)
			if err != nil {
				log.Fatalf("Error opening CSV file: %v", err)
			}
			defer file.Close()

			reader := csv.NewReader(file)
			records, err := reader.ReadAll()
			if err != nil {
				log.Fatalf("Error reading CSV file: %v", err)
			}

			// Expecting headers "date", "value"
			if len(records) < 1 {
				log.Fatalf("CSV file is empty")
			}

			headers := records[0]
			dateIdx := -1
			valueIdx := -1
			for i, header := range headers {
				lowerHeader := strings.ToLower(header)
				if lowerHeader == "date" {
					dateIdx = i
				} else if lowerHeader == "value" {
					valueIdx = i
				}
			}

			if dateIdx == -1 || valueIdx == -1 {
				log.Fatalf("CSV file must have 'date' and 'value' columns")
			}

			// Counters for feedback
			successfulImports := 0
			skippedImports := 0

			// Process each record
			for _, record := range records[1:] {
				dateStr := record[dateIdx]
				valueStr := record[valueIdx]

				// Extract only 'YYYY-MM' from 'YYYY-MM-DD'
				if len(dateStr) >= 7 {
					dateStr = dateStr[:7]
				} else {
					log.Printf("Skipping invalid date format '%s': insufficient length", dateStr)
					skippedImports++
					continue
				}

				// Parse date
				date, err := time.Parse("2006-01", dateStr)
				if err != nil {
					log.Printf("Skipping invalid date format '%s': %v", dateStr, err)
					skippedImports++
					continue
				}
				yearStr := fmt.Sprintf("%d", date.Year())
				monthStr := fmt.Sprintf("%02d", date.Month()) // Zero-pad month

				// Handle value: remove quotes and replace ',' with '.'
				valueStr = strings.Trim(valueStr, "\"")
				valueStr = strings.ReplaceAll(valueStr, ",", ".")
				value, err := strconv.ParseFloat(valueStr, 64)
				if err != nil {
					log.Printf("Skipping invalid value '%s': %v", valueStr, err)
					skippedImports++
					continue
				}

				// Initialize year if not present
				if c.Inflation == nil {
					c.Inflation = make(map[string]map[string]float64)
				}
				if _, exists := c.Inflation[yearStr]; !exists {
					c.Inflation[yearStr] = make(map[string]float64)
				}

				// Update rate
				c.Inflation[yearStr][monthStr] = value
				successfulImports++
			}

			// Save back to JSON
			err = inflation.SaveInflationData(loader.Data, *jsonFile)
			if err != nil {
				log.Fatalf("Error saving JSON data: %v", err)
			}

			fmt.Printf("Successfully imported %d records. Skipped %d records due to errors.\n", successfulImports, skippedImports)
			fmt.Printf("Successfully imported inflation rates from %s into %s for country %s with Base Year %d\n", *csvFile, *jsonFile, *country, c.BaseYear)
		}
	})

	// Command: listCountries
	app.Command("listCountries", "List all available countries", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			loader := &inflation.Loader{}
			err := loader.LoadData(*inflationList, *cacheList)
			if err != nil {
				log.Fatalf("Error loading data: %v", err)
			}

			fmt.Println("Available Countries:")
			for _, country := range loader.Data.Countries {
				fmt.Printf("- %s (Code: %s, Aliases: %v, Base Year: %d)\n", country.Name, country.Code, country.Aliases, country.BaseYear)
			}
		}
	})

	app.Action = func() {
		// Default action: display help
		app.PrintHelp()
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// parseDate parses a date string in "YYYY" or "YYYY-MM" format.
// Returns year, month (0 if not specified), error
func parseDate(dateStr string) (int, int, error) {
	if len(dateStr) == 4 {
		// "YYYY" format
		year, err := strconv.Atoi(dateStr)
		if err != nil {
			return 0, 0, err
		}
		return year, 0, nil
	} else if len(dateStr) == 7 {
		// "YYYY-MM" format
		parts := strings.Split(dateStr, "-")
		if len(parts) != 2 {
			return 0, 0, fmt.Errorf("invalid date format: %s", dateStr)
		}
		year, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid year in date: %v", err)
		}
		month, err := strconv.Atoi(parts[1])
		if err != nil || month < 1 || month > 12 {
			return 0, 0, fmt.Errorf("invalid month in date: %v", err)
		}
		return year, month, nil
	} else {
		return 0, 0, fmt.Errorf("invalid date format: %s", dateStr)
	}
}
