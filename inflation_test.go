// inflation_test.go
package inflation

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

// Epsilon for floating-point comparison
const epsilon = 1e-6

// Helper function to check if two floats are approximately equal
func floatsAlmostEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

// Helper function to create a Data instance with predefined countries and inflation rates.
func createTestData() Data {
	return Data{
		Countries: []Country{
			{
				Name:     "United States",
				Aliases:  []string{"US", "USA"},
				Code:     "US",
				BaseYear: 2015,
				Inflation: map[string]map[string]float64{
					"2015": {
						"01": 0.1,
						"02": 0.2,
						"03": 0.3,
						"04": 0.2,
						"05": 0.1,
						"06": 0.3,
						"07": 0.2,
						"08": 0.1,
						"09": 0.3,
						"10": 0.2,
						"11": 0.1,
						"12": 0.3,
					},
					"2018": {
						"01": 0.2,
						"02": 0.3,
						"03": 0.4,
						"04": 0.3,
						"05": 0.2,
						"06": 0.4,
						"07": 0.3,
						"08": 0.2,
						"09": 0.4,
						"10": 0.3,
						"11": 0.2,
						"12": 0.4,
					},
					"2016": {
						"01": 0.15,
						"02": 0.25,
						"03": 0.35,
						"04": 0.25,
						"05": 0.15,
						"06": 0.35,
						"07": 0.25,
						"08": 0.15,
						"09": 0.35,
						"10": 0.25,
						"11": 0.15,
						"12": 0.35,
					},
				},
			},
			{
				Name:     "Germany",
				Aliases:  []string{"DE", "GER"},
				Code:     "DE",
				BaseYear: 2015,
				Inflation: map[string]map[string]float64{
					"2015": {
						"01": 0.05,
						"02": 0.07,
						"03": 0.06,
						"04": 0.08,
						"05": 0.07,
						"06": 0.09,
						"07": 0.06,
						"08": 0.05,
						"09": 0.07,
						"10": 0.06,
						"11": 0.05,
						"12": 0.07,
					},
					"2018": {
						"01": 0.1,
						"02": 0.12,
						"03": 0.11,
						"04": 0.13,
						"05": 0.12,
						"06": 0.14,
						"07": 0.11,
						"08": 0.1,
						"09": 0.12,
						"10": 0.11,
						"11": 0.1,
						"12": 0.12,
					},
				},
			},
		},
	}
}

func TestGetCountry(t *testing.T) {
	data := createTestData()

	tests := []struct {
		query       string
		expected    string
		expectError bool
	}{
		{"United States", "United States", false},
		{"US", "United States", false},
		{"USA", "United States", false},
		{"Germany", "Germany", false},
		{"DE", "Germany", false},
		{"GER", "Germany", false},
		{"France", "", true}, // Non-existent country
	}

	for _, tt := range tests {
		country, err := data.GetCountry(tt.query)
		if tt.expectError {
			if err == nil {
				t.Errorf("Expected error for query '%s', but got none", tt.query)
			}
		} else {
			if err != nil {
				t.Errorf("Did not expect error for query '%s', but got: %v", tt.query, err)
			} else if country.Name != tt.expected {
				t.Errorf("For query '%s', expected country '%s', got '%s'", tt.query, tt.expected, country.Name)
			}
		}
	}
}

func TestYearInflation(t *testing.T) {
	data := createTestData()

	tests := []struct {
		country     string
		year        int
		month       int
		expected    float64
		expectError bool
	}{
		// Average inflation rate for 2015
		{"US", 2015, 0, 0.2, false},
		// Specific month inflation rate for 2015-03
		{"US", 2015, 3, 0.3, false},
		// Specific month inflation rate for 2018-06
		{"US", 2018, 6, 0.4, false},
		// Non-existent year
		{"US", 2020, 0, 0.0, true},
		// Non-existent month
		{"US", 2015, 13, 0.0, true},
		// Non-existent country
		{"France", 2015, 0, 0.0, true},
		// Average inflation rate for Germany
		{"Germany", 2015, 0, 0.065, false}, // Corrected from 0.07 to 0.065
		// Specific month inflation rate for Germany-04
		{"Germany", 2015, 4, 0.08, false},
	}

	for _, tt := range tests {
		rate, err := data.YearInflation(tt.country, tt.year, tt.month)
		if tt.expectError {
			if err == nil {
				t.Errorf("Expected error for country='%s', year=%d, month=%d, but got none", tt.country, tt.year, tt.month)
			}
		} else {
			if err != nil {
				t.Errorf("Did not expect error for country='%s', year=%d, month=%d, but got: %v", tt.country, tt.year, tt.month, err)
			} else if !floatsAlmostEqual(rate, tt.expected) {
				t.Errorf("For country='%s', year=%d, month=%d, expected rate=%.6f, got=%.6f", tt.country, tt.year, tt.month, tt.expected, rate)
			}
		}
	}
}

func TestCompareInflation(t *testing.T) {
	data := createTestData()

	tests := []struct {
		country         string
		fromYear        int
		fromMonth       int
		toYear          int
		toMonth         int
		price           float64
		expectedPrice   float64
		expectedInflate float64
		expectError     bool
	}{
		// Basic comparison between two years
		{"US", 2015, 0, 2018, 0, 35, 35 * (0.3 / 0.2), (0.3/0.2 - 1) * 100, false}, // 35 *1.5=52.5, 50%
		// Comparison between specific months
		{"US", 2015, 6, 2018, 6, 35, 35 * (0.4 / 0.3), (0.4/0.3 - 1) * 100, false}, // 35*1.3333≈46.6667, 33.3333%
		// From month to year
		{"US", 2015, 6, 2018, 0, 35, 35 * (0.3 / 0.3), (0.3/0.3 - 1) * 100, false}, // 35*1=35, 0%
		// From year to month
		{"US", 2015, 0, 2018, 6, 35, 35 * (0.4 / 0.2), (0.4/0.2 - 1) * 100, false}, // 35*2=70, 100%
		// Non-existent from date
		{"US", 2020, 0, 2018, 0, 35, 0.0, 0.0, true},
		// Non-existent to date
		{"US", 2015, 0, 2020, 0, 35, 0.0, 0.0, true},
		// Non-existent country
		{"France", 2015, 0, 2018, 0, 35, 0.0, 0.0, true},
		// Specific months with known rates for Germany
		{"Germany", 2015, 4, 2018, 4, 100, 100 * (0.13 / 0.08), (0.13/0.08 - 1) * 100, false}, // ≈162.5, 62.5%
	}

	for _, tt := range tests {
		newPrice, cumulativeRate, err := data.CompareInflation(tt.country, tt.fromYear, tt.fromMonth, tt.toYear, tt.toMonth, tt.price)
		if tt.expectError {
			if err == nil {
				t.Errorf("Expected error for compareInflation with country='%s', fromYear=%d, fromMonth=%d, toYear=%d, toMonth=%d, price=%.2f, but got none",
					tt.country, tt.fromYear, tt.fromMonth, tt.toYear, tt.toMonth, tt.price)
			}
		} else {
			if err != nil {
				t.Errorf("Did not expect error for compareInflation with country='%s', fromYear=%d, fromMonth=%d, toYear=%d, toMonth=%d, price=%.2f, but got: %v",
					tt.country, tt.fromYear, tt.fromMonth, tt.toYear, tt.toMonth, tt.price, err)
			} else {
				// Allow a small margin for floating point comparison
				if !floatsAlmostEqual(newPrice, tt.expectedPrice) {
					t.Errorf("For compareInflation with country='%s', fromYear=%d, fromMonth=%d, toYear=%d, toMonth=%d, price=%.2f, expected newPrice=%.2f, got=%.2f",
						tt.country, tt.fromYear, tt.fromMonth, tt.toYear, tt.toMonth, tt.price, tt.expectedPrice, newPrice)
				}
				if !floatsAlmostEqual(cumulativeRate, tt.expectedInflate) {
					t.Errorf("For compareInflation with country='%s', fromYear=%d, fromMonth=%d, toYear=%d, toMonth=%d, price=%.2f, expected cumulativeRate=%.2f%%, got=%.2f%%",
						tt.country, tt.fromYear, tt.fromMonth, tt.toYear, tt.toMonth, tt.price, tt.expectedInflate, cumulativeRate)
				}
			}
		}
	}
}

func TestCompareInflationWithBaseYear(t *testing.T) {
	data := createTestData()

	tests := []struct {
		name          string
		country       string
		targetYear    int
		targetMonth   int
		price         float64
		expectedPrice float64
		expectError   bool
	}{
		// Basic comparison relative to BaseYear
		{
			name:          "US average year comparison",
			country:       "US",
			targetYear:    2018,
			targetMonth:   0,
			price:         35,
			expectedPrice: 35 * (0.3 / 0.2), // 35 *1.5 =52.5
			expectError:   false,
		},
		// Specific month comparison
		{
			name:          "US specific month comparison",
			country:       "US",
			targetYear:    2018,
			targetMonth:   6,
			price:         35,
			expectedPrice: 35 * (0.4 / 0.2), // 35 *2.0 =70.0
			expectError:   false,
		},
		// Non-existent target date
		{
			name:          "US non-existent target date",
			country:       "US",
			targetYear:    2020,
			targetMonth:   0,
			price:         35,
			expectedPrice: 0.0,
			expectError:   true,
		},
		// Non-existent country
		{
			name:          "Non-existent country",
			country:       "France",
			targetYear:    2018,
			targetMonth:   0,
			price:         35,
			expectedPrice: 0.0,
			expectError:   true,
		},
		// Country without BaseYear
		{
			name:          "Country without BaseYear",
			country:       "Spain",
			targetYear:    2018,
			targetMonth:   6,
			price:         50,
			expectedPrice: 0.0,
			expectError:   true,
		},
		// Germany specific month comparison
		{
			name:          "Germany specific month comparison",
			country:       "Germany",
			targetYear:    2018,
			targetMonth:   4,
			price:         100,
			expectedPrice: 100 * (0.13 / 0.065), // ≈200.0
			expectError:   false,
		},
	}

	// Adding Spain without BaseYear to the data
	data.Countries = append(data.Countries, Country{
		Name:      "Spain",
		Aliases:   []string{"ES"},
		Code:      "ES",
		BaseYear:  0, // No BaseYear
		Inflation: map[string]map[string]float64{},
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newPrice, err := data.CompareInflationWithBaseYear(tt.country, tt.targetYear, tt.targetMonth, tt.price)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for CompareInflationWithBaseYear with country='%s', targetYear=%d, targetMonth=%d, price=%.2f, but got none",
						tt.country, tt.targetYear, tt.targetMonth, tt.price)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error for CompareInflationWithBaseYear with country='%s', targetYear=%d, targetMonth=%d, price=%.2f, but got: %v",
						tt.country, tt.targetYear, tt.targetMonth, tt.price, err)
				} else if !floatsAlmostEqual(newPrice, tt.expectedPrice) {
					t.Errorf("For CompareInflationWithBaseYear with country='%s', targetYear=%d, targetMonth=%d, price=%.2f, expected newPrice=%.2f, got=%.2f",
						tt.country, tt.targetYear, tt.targetMonth, tt.price, tt.expectedPrice, newPrice)
				}
			}
		})
	}
}

func TestCompareInflation_EdgeCases(t *testing.T) {
	data := createTestData()

	tests := []struct {
		name        string
		country     string
		fromYear    int
		fromMonth   int
		toYear      int
		toMonth     int
		price       float64
		expectError bool
	}{
		{
			name:        "Invalid fromMonth",
			country:     "US",
			fromYear:    2015,
			fromMonth:   13, // Invalid month
			toYear:      2018,
			toMonth:     6,
			price:       35,
			expectError: true,
		},
		{
			name:        "Invalid toMonth",
			country:     "US",
			fromYear:    2015,
			fromMonth:   6,
			toYear:      2018,
			toMonth:     0, // Valid (average)
			price:       35,
			expectError: false,
		},
		{
			name:        "Non-integer month",
			country:     "US",
			fromYear:    2015,
			fromMonth:   -1, // Invalid month
			toYear:      2018,
			toMonth:     6,
			price:       35,
			expectError: true,
		},
		{
			name:        "Zero price",
			country:     "US",
			fromYear:    2015,
			fromMonth:   6,
			toYear:      2018,
			toMonth:     6,
			price:       0.0,   // Zero price
			expectError: false, // Depending on business logic, this might be allowed
		},
		{
			name:        "Negative price",
			country:     "US",
			fromYear:    2015,
			fromMonth:   6,
			toYear:      2018,
			toMonth:     6,
			price:       -100.0, // Negative price
			expectError: false,  // Depending on business logic, this might be allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := data.CompareInflation(tt.country, tt.fromYear, tt.fromMonth, tt.toYear, tt.toMonth, tt.price)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for test '%s', but got none", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error for test '%s', but got: %v", tt.name, err)
				}
			}
		})
	}
}

func TestCompareInflationWithBaseYear_EdgeCases(t *testing.T) {
	data := createTestData()

	tests := []struct {
		name        string
		country     string
		targetYear  int
		targetMonth int
		price       float64
		expectError bool
	}{
		{
			name:        "Country without BaseYear",
			country:     "Spain",
			targetYear:  2018,
			targetMonth: 6,
			price:       50.0,
			expectError: true,
		},
		{
			name:        "Target date before BaseYear",
			country:     "US",
			targetYear:  2014,
			targetMonth: 12,
			price:       100.0,
			expectError: true, // Assuming no data for 2014
		},
		{
			name:        "Invalid targetMonth",
			country:     "US",
			targetYear:  2018,
			targetMonth: 13, // Invalid month
			price:       100.0,
			expectError: true,
		},
		{
			name:        "Valid comparison with specific month",
			country:     "US",
			targetYear:  2018,
			targetMonth: 6,
			price:       35.0,
			expectError: false,
		},
		{
			name:        "Valid comparison with average year",
			country:     "US",
			targetYear:  2018,
			targetMonth: 0,
			price:       35.0,
			expectError: false,
		},
		{
			name:        "Valid comparison for Germany",
			country:     "Germany",
			targetYear:  2018,
			targetMonth: 4,
			price:       100.0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newPrice, err := data.CompareInflationWithBaseYear(tt.country, tt.targetYear, tt.targetMonth, tt.price)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for CompareInflationWithBaseYear with country='%s', targetYear=%d, targetMonth=%d, price=%.2f, but got none",
						tt.country, tt.targetYear, tt.targetMonth, tt.price)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error for CompareInflationWithBaseYear with country='%s', targetYear=%d, targetMonth=%d, price=%.2f, but got: %v",
						tt.country, tt.targetYear, tt.targetMonth, tt.price, err)
				} else {
					var expectedPrice float64
					switch tt.name {
					case "Valid comparison with specific month":
						// US, 2018-6: 0.4 / 0.2 = 2.0; 35 *2.0=70
						expectedPrice = 35.0 * (0.4 / 0.2) // 70.0
					case "Valid comparison with average year":
						// US, 2018-0: 0.3 / 0.2 =1.5; 35 *1.5=52.5
						expectedPrice = 35.0 * (0.3 / 0.2) // 52.5
					case "Valid comparison for Germany":
						// Germany, 2018-4: 0.13 /0.065≈2.0; 100 *2.0=200.0
						expectedPrice = 100.0 * (0.13 / 0.065) // 200.0
					default:
						t.Fatalf("Unexpected test case name: %s", tt.name)
					}
					if !floatsAlmostEqual(newPrice, expectedPrice) {
						t.Errorf("For CompareInflationWithBaseYear with country='%s', targetYear=%d, targetMonth=%d, price=%.2f, expected newPrice=%.2f, got=%.2f",
							tt.country, tt.targetYear, tt.targetMonth, tt.price, expectedPrice, newPrice)
					}
				}
			}
		})
	}
}

func TestLoadInflationData(t *testing.T) {
	// Create temporary JSON file
	tempFile, err := os.CreateTemp("", "test_inflation_data_*.json")
	if err != nil {
		t.Fatalf("Failed to create temporary JSON file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up

	// Create mock JSON data
	mockData := createTestData()
	jsonBytes, err := json.Marshal(mockData)
	if err != nil {
		t.Fatalf("Failed to marshal mock data: %v", err)
	}

	// Write mock data to temporary file
	_, err = tempFile.Write(jsonBytes)
	if err != nil {
		t.Fatalf("Failed to write mock data to temporary file: %v", err)
	}
	tempFile.Close()

	// Test loading from local file
	loader := &Loader{}
	err = loader.LoadData(tempFile.Name(), false)
	if err != nil {
		t.Errorf("Failed to load inflation data from local file: %v", err)
	}

	// Verify loaded data
	if len(loader.Data.Countries) != 2 {
		t.Errorf("Expected 2 countries, got %d", len(loader.Data.Countries))
	}

	// Test loading from invalid file
	loaderInvalid := &Loader{}
	err = loaderInvalid.LoadData("non_existent_file.json", false)
	if err == nil {
		t.Errorf("Expected error when loading from non-existent file, but got none")
	}

	// Optionally, test loading from a URL using a mock server or similar approach.
	// This is more involved and typically requires additional packages or setup.
}

func TestSaveInflationData(t *testing.T) {
	data := createTestData()

	// Create temporary JSON file
	tempFile, err := os.CreateTemp("", "test_save_inflation_data_*.json")
	if err != nil {
		t.Fatalf("Failed to create temporary JSON file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up

	// Save data to temporary file
	err = SaveInflationData(data, tempFile.Name())
	if err != nil {
		t.Errorf("Failed to save inflation data: %v", err)
	}

	// Read back the file and verify contents
	savedData := Data{}
	file, err := os.Open(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to open saved inflation data file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&savedData)
	if err != nil {
		t.Errorf("Failed to decode saved inflation data: %v", err)
	}

	// Compare original and saved data
	originalBytes, _ := json.Marshal(data)
	savedBytes, _ := json.Marshal(savedData)
	if string(originalBytes) != string(savedBytes) {
		t.Errorf("Saved data does not match original data.\nOriginal: %s\nSaved: %s", string(originalBytes), string(savedBytes))
	}
}
