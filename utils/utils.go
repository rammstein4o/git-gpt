// ---
// Copyright © 2023 Radoslav Salov <rado.salov@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
// ---

package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

var (
	binaryExtensions = []string{
		".ai",
		".bmp",
		".eps",
		".gif",
		".ico",
		".jng",
		".jp2",
		".jpg",
		".jpeg",
		".jpx",
		".jxr",
		".pdf",
		".png",
		".psb",
		".psd",
		".svgz",
		".tif",
		".tiff",
		".wbmp",
		".webp",
		".kar",
		".m4a",
		".mid",
		".midi",
		".mp3",
		".ogg",
		".ra",
		".3gpp",
		".3gp",
		".as",
		".asf",
		".asx",
		".fla",
		".flv",
		".m4v",
		".mng",
		".mov",
		".mp4",
		".mpeg",
		".mpg",
		".ogv",
		".swc",
		".swf",
		".webm",
		".7z",
		".gz",
		".jar",
		".rar",
		".tar",
		".zip",
		".ttf",
		".eot",
		".otf",
		".woff",
		".woff2",
		".exe",
		".pyc",
	}
)

// IsCommandAvailable checks whether a command is available in the PATH.
func IsCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// ConvertToMap converts a slice of strings to a map.
func ConvertToMap(args []string) Data {
	m := make(Data)
	for _, arg := range args {
		kv := strings.SplitN(arg, "=", 2)
		if len(kv) == 2 {
			m[kv[0]] = kv[1]
		}
	}
	return m
}

// IsFile returns true if given path is a file,
// or returns false when it's a directory or does not exist.
func IsFile(filePath string) bool {
	f, e := os.Stat(filePath)
	if e != nil {
		return false
	}
	return !f.IsDir()
}

// IsDir returns true if given path is a directory,
// or returns false when it's a file or does not exist.
func IsDir(dir string) bool {
	f, e := os.Stat(dir)
	if e != nil {
		return false
	}
	return f.IsDir()
}

func ReadFile(filePath string) (string, error) {
	if !IsFile(filePath) {
		return "", fmt.Errorf("file %s does not exist", filePath)
	}

	// Open a file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Use io.ReadAll to read from the file
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func findGitDir(dir string) string {
	// set maxDepth to 1000
	maxDepth := 1000
	for i := 0; i < maxDepth; i++ {
		// Check if the current directory contains a .git folder
		gitDir := filepath.Join(dir, ".git")
		_, err := os.Stat(gitDir)
		if err == nil {
			return dir
		}

		// Move to the parent directory
		parentDir := filepath.Dir(dir)
		// Check if we have reached the root directory
		if parentDir == dir {
			break
		}
		dir = parentDir
	}

	return ""
}

func CwdToGitRoot() error {
	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// Find the closest parent directory with a .git folder
	gitDir := findGitDir(currentDir)
	if gitDir == "" {
		return fmt.Errorf("no .git directory found in any parent directory")
	}

	// Change the current working directory to the one containing .git
	err = os.Chdir(gitDir)
	if err != nil {
		return err
	}

	return nil
}

func IsBinaryFile(fileName string) bool {
	extension := strings.ToLower(filepath.Ext(fileName))
	for _, ext := range binaryExtensions {
		if ext == extension {
			return true
		}
	}
	return false
}

func SplitText(text string, chunkSize int) []string {
	var chunks []string

	// Split the text into words
	words := strings.Fields(text)

	// Initialize variables to track current chunk and remaining space
	currentChunk := ""
	remainingSpace := chunkSize

	// Iterate through words and add them to the current chunk until it reaches the chunkSize
	for _, word := range words {
		wordLength := utf8.RuneCountInString(word) + 1 // Add 1 for space

		// If adding the current word doesn't exceed the remaining space, add it to the chunk
		if wordLength <= remainingSpace {
			currentChunk += word + " "
			remainingSpace -= wordLength
		} else {
			// If adding the word exceeds the remaining space, start a new chunk
			chunks = append(chunks, strings.TrimSpace(currentChunk))
			currentChunk = word + " "
			remainingSpace = chunkSize - wordLength
		}
	}

	// Add the last chunk
	chunks = append(chunks, strings.TrimSpace(currentChunk))

	return chunks
}
