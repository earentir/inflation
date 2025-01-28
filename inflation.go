// inflation/inflation.go
package inflation

import (
	"fmt"
	"strconv"
	"strings"
)

// GetCountry retrieves a country by name, alias, or code.
func (d *Data) GetCountry(query string) (*Country, error) {
	query = strings.ToLower(query)
	for i, country := range d.Countries {
		if strings.ToLower(country.Name) == query || strings.ToLower(country.Code) == query {
			return &d.Countries[i], nil // Return pointer to the actual country in the slice
		}
		for _, alias := range country.Aliases {
			if strings.ToLower(alias) == query {
				return &d.Countries[i], nil
			}
		}
	}
	return nil, fmt.Errorf("country '%s' not found", query)
}

// GetAvailableYears returns a list of available years for a country.
func (c *Country) GetAvailableYears() []int {
	years := make([]int, 0, len(c.Inflation))
	for year := range c.Inflation {
		y, _ := strconv.Atoi(year)
		years = append(years, y)
	}
	return years
}

// GetFirstDate returns the earliest year and month for a country.
func (c *Country) GetFirstDate() (year int, month int) {
	for yearStr, months := range c.Inflation {
		y, _ := strconv.Atoi(yearStr)
		if year == 0 || y < year {
			year = y
			for monthStr := range months {
				m, _ := strconv.Atoi(monthStr)
				if month == 0 || m < month {
					month = m
				}
			}
		}
	}
	return year, month
}

// GetLastDate returns the latest year and month for a country.
func (c *Country) GetLastDate() (year int, month int) {
	for yearStr, months := range c.Inflation {
		y, _ := strconv.Atoi(yearStr)
		if y > year {
			year = y
			for monthStr := range months {
				m, _ := strconv.Atoi(monthStr)
				if m > month {
					month = m
				}
			}
		}
	}
	return year, month
}

// YearInflation returns the inflation rate for a specific country and date.
// If month is 0, it returns the average inflation rate for the year.
// If month is between 1 and 12, it returns the rate for that specific month.
func (d *Data) YearInflation(country string, year int, month int) (float64, error) {
	c, err := d.GetCountry(country)
	if err != nil {
		return 0, err
	}
	yearStr := fmt.Sprintf("%d", year)
	yearData, exists := c.Inflation[yearStr]
	if !exists {
		return 0, fmt.Errorf("inflation data for year %d not found for country '%s'", year, country)
	}
	if month == 0 {
		// Calculate average of all months
		var sum float64
		var count int
		for _, rate := range yearData {
			sum += rate
			count++
		}
		if count == 0 {
			return 0, fmt.Errorf("no monthly data available for year %d in country '%s'", year, country)
		}
		return sum / float64(count), nil
	} else if month >= 1 && month <= 12 {
		monthStr := fmt.Sprintf("%02d", month) // Ensure monthStr is zero-padded
		rate, exists := yearData[monthStr]
		if !exists {
			return 0, fmt.Errorf("inflation data for %d-%02d not found for country '%s'", year, month, country)
		}
		return rate, nil
	} else {
		return 0, fmt.Errorf("invalid month: %d", month)
	}
}

// CompareInflation calculates the equivalent price adjusted for inflation between two dates for a country.
// Returns both the new price and the cumulative rate of inflation.
func (d *Data) CompareInflation(country string, fromYear, fromMonth int, toYear, toMonth int, price float64) (float64, float64, error) {
	fromRate, err := d.YearInflation(country, fromYear, fromMonth)
	if err != nil {
		return 0, 0, err
	}
	toRate, err := d.YearInflation(country, toYear, toMonth)
	if err != nil {
		return 0, 0, err
	}

	// Calculate inflation factor
	// Inflation Factor = toRate / fromRate
	inflationFactor := toRate / fromRate
	newPrice := price * inflationFactor

	// Calculate cumulative rate of inflation
	// Cumulative Rate (%) = (inflationFactor - 1) * 100
	cumulativeRate := (inflationFactor - 1) * 100

	return newPrice, cumulativeRate, nil
}

// CompareInflationWithBaseYear calculates the equivalent price adjusted for inflation relative to the BaseYear.
func (d *Data) CompareInflationWithBaseYear(country string, targetYear, targetMonth int, price float64) (float64, error) {
	c, err := d.GetCountry(country)
	if err != nil {
		return 0, err
	}

	// Get the BaseYear
	baseYear := c.BaseYear
	if baseYear == 0 {
		return 0, fmt.Errorf("base year not set for country '%s'", country)
	}

	// Get inflation rate for BaseYear (average of the year)
	baseRate, err := d.YearInflation(country, baseYear, 0)
	if err != nil {
		return 0, fmt.Errorf("error fetching BaseYear inflation rate: %v", err)
	}

	// Get inflation rate for target date
	targetRate, err := d.YearInflation(country, targetYear, targetMonth)
	if err != nil {
		return 0, fmt.Errorf("error fetching target inflation rate: %v", err)
	}

	// Calculate inflation factor relative to BaseYear
	inflationFactor := targetRate / baseRate
	newPrice := price * inflationFactor

	return newPrice, nil
}
