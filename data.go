// inflation/data.go
package inflation

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
)

// Data holds the inflation rates for multiple countries.
type Data struct {
	Countries []Country `json:"countries"`
}

// Country represents a country's inflation information.
type Country struct {
	Name      string                        `json:"name"`
	Aliases   []string                      `json:"aliases"`
	Code      string                        `json:"code"`
	BaseYear  int                           `json:"base_year"` // HICP Base Year
	Inflation map[string]map[string]float64 `json:"inflation"` // Year -> Month -> Rate
}

// Loader is responsible for loading inflation data.
type Loader struct {
	Data       Data
	dataLoaded bool
}

// LoadData loads the inflation data from the provided source.
// It accepts a 'cache' boolean to decide whether to cache the data if fetched from a URL.
func (l *Loader) LoadData(source string, cache bool) error {
	data, err := LoadInflationData(source, cache)
	if err != nil {
		return err
	}
	l.Data = data
	l.dataLoaded = true
	return nil
}

// LoadInflationData loads inflation data from a local file or a URL.
func LoadInflationData(source string, cache bool) (Data, error) {
	var data Data

	if isURL(source) {
		// Fetch from URL
		resp, err := http.Get(source)
		if err != nil {
			return data, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return data, errors.New("failed to fetch inflation data from URL")
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return data, err
		}

		// Optionally cache the data
		if cache {
			err = os.WriteFile("inflationratelist.json", body, 0644)
			if err != nil {
				return data, err
			}
		}

		err = json.Unmarshal(body, &data)
		if err != nil {
			return data, err
		}
	} else {
		// Load from local file
		file, err := os.Open(source)
		if err != nil {
			return data, err
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		err = decoder.Decode(&data)
		if err != nil {
			return data, err
		}
	}

	return data, nil
}

// isURL checks if the source string is a URL.
func isURL(source string) bool {
	return strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://")
}

// SaveInflationData saves the Data back to the specified JSON file.
func SaveInflationData(data Data, filePath string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}
