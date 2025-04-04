package tle

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	// TLE_URL = "https://celestrak.org/NORAD/elements/weather.txt"
	TLE_FILE = "weather.txt"
)

// Checks if weather.txt and weather.txt.timestamp are valid
func CheckTimestamp() bool {
	// Check if weather.txt itself exists
	_, err := os.ReadFile(TLE_FILE)
	if err != nil {
		return false
	}

	// Now check if the timestamp is valid
	data, err := os.ReadFile(TLE_FILE + ".timestamp")
	if err != nil {
		return false
	}

	// Parse timestamp
	ts, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return false
	}

	return time.Since(ts) < 24 * time.Hour
}

func FetchTLEs(url string) error {
	if CheckTimestamp() {
		fmt.Println("Using cached TLE data.")
		return nil
	}

	// We need to download + create timestamp
	fmt.Println("Cached TLE data is missing or outdated. Fetching the latest TLE data...")

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Save the file
	file, err := os.Create(TLE_FILE)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	// Create timestamp file
	os.WriteFile(TLE_FILE + ".timestamp", []byte(time.Now().Format(time.RFC3339)), 0644)
	fmt.Println("TLE data updated successfully!")

	return nil
}

func LoadTLEs() (map[string][2]string, error) {
	file, err := os.Open(TLE_FILE)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	satellites := make(map[string][2]string)

	var name, line1, line2 string
	for scanner.Scan() {
		if name == "" {
			name = strings.TrimSpace(scanner.Text())
		} else if line1 == "" {
			line1 = scanner.Text()
		} else {
			line2 = scanner.Text()
			satellites[name] = [2]string{line1, line2}
			name, line1, line2 = "", "", ""
		}
	}

	return satellites, nil
}

