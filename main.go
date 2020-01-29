package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func main() {
	inputDir := flag.String("target-dir", ".", "Directory where files to rename are found. Defaults to '.'")
	flag.Parse()
	files, err := ioutil.ReadDir(*inputDir)
	if err != nil {
		panic(fmt.Errorf("I was unable to read the directory: `%s`. I got this error:\n%v", *inputDir, err))
	}

	newDir := filepath.Join(*inputDir, "uncategorized")
	err = os.MkdirAll(newDir, 0755)
	if err != nil {
		panic(fmt.Errorf("I couldn't create the directory for the non-matching files. Got this error:\n%v", err))
	}

	nonMatches := renameFiles(files, *inputDir)
	moveNonMatches(nonMatches, *inputDir, newDir)
}

func moveNonMatches(files []string, inputDir, newDir string) {
	for _, badMatch := range files {
		oldPath := filepath.Join(inputDir, badMatch)
		newPath := filepath.Join(newDir, badMatch)
		err := os.Rename(oldPath, newPath)
		if err != nil {
			fmt.Printf("I couldn't rename this file: `%s`. Got this error:\n%v", badMatch, err)
		}
	}
}

func renameFiles(files []os.FileInfo, inputDir string) []string {
	var nonMatches []string
	for _, file := range files {
		matched, err := match(file.Name())
		if err != nil {
			fmt.Printf("`%s` didn't match the pattern I was looking for\n", file)
		}

		if !matched {
			nonMatches = append(nonMatches, file.Name())
			continue
		}

		oldPath := filepath.Join(inputDir, file.Name())
		nameAndExt := strings.Split(file.Name(), ".")
		dateTaken := strings.Split(nameAndExt[0], "_")[0]
		date, err := time.Parse("20060102", dateTaken)
		if err != nil {
			panic(fmt.Errorf("I couldn't parse the date `%s` correctly. Got this error\n %v", dateTaken, err))
		}

		ext := nameAndExt[1]
		newName := fmt.Sprintf("%s--%s.%s", date.Format("January_02_2006"), strings.Split(nameAndExt[0], "_")[1], ext)
		newPath := filepath.Join(inputDir, newName)
		err = os.Rename(oldPath, newPath)
		if err != nil {
			panic(fmt.Errorf("I was unable to rename `%s`. Got this error:\n%v", file.Name(), err))
		}
	}
	return nonMatches
}

func match(filename string) (bool, error) {
	isMatch, err := regexp.Match(`^\d{8}_\d+.jpg$`, []byte(filename))
	if err != nil {
		return false, err
	}
	return isMatch, nil
}
