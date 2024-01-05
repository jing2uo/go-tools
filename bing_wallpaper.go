package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
)

// BingAPIResponse represents the JSON response from Bing API
type BingAPIResponse struct {
	Images []struct {
		URL string `json:"url"`
	} `json:"images"`
}

func getBingWallpaperURL() (string, error) {
	apiURL := "https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=1&mkt=en-US"
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data BingAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if len(data.Images) == 0 {
		return "", fmt.Errorf("no images found in the response")
	}

	return "https://www.bing.com" + data.Images[0].URL, nil
}

func extractFileName(fileURL string) (string, error) {
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", err
	}

	// Extract the query parameters
	params := parsedURL.Query()
	rf := params.Get("rf")
	if rf == "" {
		return "", fmt.Errorf("parameter 'rf' not found in the URL")
	}

	return rf, nil
}

func downloadWallpaper(bingURL, dirPath string) (string, error) {
	resp, err := http.Get(bingURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fileName, err := extractFileName(bingURL)
	if err != nil {
		return "", err
	}
	filePath := filepath.Join(dirPath, fileName)

	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return filePath, nil
}
func main() {
	usr, err := user.Current()
	if err != nil {
		fmt.Printf("Error getting current user: %s\n", err)
		os.Exit(1)
	}
	defaultDirPath := filepath.Join(usr.HomeDir, "Pictures")

	// Define and parse the command line flags
	dirPath := flag.String("o", defaultDirPath, "Directory to save the wallpaper")
	flag.Parse()

	// Check if directory exists and is readable
	dirInfo, err := os.Stat(*dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Provided directory does not exist: %s", *dirPath)
			os.Exit(1)
		}
		fmt.Printf("Error accessing provided directory: %s", err)
		os.Exit(1)
	}

	if !dirInfo.IsDir() {
		fmt.Printf("Provided path is not a directory: %s", *dirPath)
		os.Exit(1)
	}

	// Check for read permission
	if dirInfo.Mode().Perm()&(1<<(uint(7))) == 0 {
		fmt.Printf("Provided directory is not readable: %s", *dirPath)
		os.Exit(1)
	}

	wallpaperURL, err := getBingWallpaperURL()
	if err != nil {
		fmt.Printf("Error fetching wallpaper URL: %s\n", err)
		os.Exit(1)
	}

	filePath, err := downloadWallpaper(wallpaperURL, *dirPath)
	if err != nil {
		fmt.Printf("Error downloading wallpaper: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Bing wallpaper download successfully: %s\n", filePath)
}
