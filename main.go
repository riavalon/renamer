package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func main() {
	inputDir := flag.String("target-dir", ".", "Directory where files to rename are found. Defaults to '.'")
	shouldRevert := flag.Bool("revert", false, "Use manifest file in given `target-dir` to revert renaming changes.")
	flag.Parse()
	files, err := ioutil.ReadDir(*inputDir)
	if err != nil {
		panic(fmt.Errorf("I was unable to read the directory: `%s`. I got this error:\n%v", *inputDir, err))
	}

	manifestFile := getManifestFile(*shouldRevert, *inputDir)
	defer manifestFile.Close()

	uncategorizedDir := filepath.Join(*inputDir, "uncategorized")
	setUncategorizedDir(*shouldRevert, uncategorizedDir)

	renameOrRevertFiles(*shouldRevert, files, *inputDir, uncategorizedDir, manifestFile)

	err = manifestFile.Sync()
	if err != nil {
		panic(fmt.Errorf("I couldn't sync the manifest file after writing.. got this error:\n%v", err))
	}
}

func renameOrRevertFiles(isRevertMode bool, files []os.FileInfo, inputDir, uncategorizedDir string, manifestFile *os.File) {
	if isRevertMode {
		records := getManifestValues(manifestFile)
		for _, record := range records {
			oldPath, newPath := record[0], record[1]
			// err := os.Rename(newPath, oldPath)
			fmt.Printf("Renaming `%s` to `%s`\n", newPath, oldPath)
			// if err != nil {
			// 	panic(fmt.Errorf("I ran into a problem while reverting file names.\n%v", err))
			// }
		}

		// os.Remove(uncategorizedDir)
		fmt.Println("Removing: ", uncategorizedDir)
		fmt.Println("Removing: ", manifestFile.Name())

		// grab oldPath and newPath
		// perform an os.Rename from the newPath to the old Path.
		// Clean up uncategorized folder and delete manifest file
	} else {
		nonMatches := renameFiles(files, inputDir, manifestFile)
		moveNonMatches(nonMatches, inputDir, uncategorizedDir, manifestFile)
	}
}

func getManifestValues(f *os.File) [][]string {
	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		panic(fmt.Errorf("I couldn't parse the CSV file into a 2D string slice. Got this error:\n%v", err))
	}
	return records
}

func setUncategorizedDir(isRevertMode bool, dirPath string) {
	if !isRevertMode {
		err := os.MkdirAll(dirPath, 0666)
		if err != nil {
			panic(fmt.Errorf("I couldn't create the directory for the non-matching files. Got this error:\n%v", err))
		}
	}
}

func getManifestFile(isRevertMode bool, inputDir string) *os.File {
	manifestPath := filepath.Join(inputDir, "manifest.csv")
	if isRevertMode {
		file, err := os.OpenFile(manifestPath, os.O_RDWR, 0666)
		if err != nil {
			panic(fmt.Errorf("I ran into an error trying to open the manifest file. Here's the error:\n%v", err))
		}
		return file
	}
	file, err := os.OpenFile(manifestPath, os.O_CREATE, 0666)
	if err != nil {
		panic(fmt.Errorf("I ran into an error trying to create the manifest file. Here's the error:\n%v", err))
	}
	return file
}

func moveNonMatches(files []string, inputDir, newDir string, manifestFile io.Writer) {
	csvWriter := csv.NewWriter(manifestFile)
	defer csvWriter.Flush()
	for _, badMatch := range files {
		oldPath := filepath.Join(inputDir, badMatch)
		newPath := filepath.Join(newDir, badMatch)
		err := os.Rename(oldPath, newPath)
		if err != nil {
			fmt.Printf("I couldn't rename this file: `%s`. Got this error:\n%v", badMatch, err)
		}
		csvField := []string{oldPath, newPath}
		err = csvWriter.Write(csvField)
		if err != nil {
			panic(fmt.Errorf("I couldn't write to the CSV file. Got this error:\n%v", err))
		}
	}
}

func renameFiles(files []os.FileInfo, inputDir string, manifestFile io.Writer) []string {
	var nonMatches []string
	csvWriter := csv.NewWriter(manifestFile)
	defer csvWriter.Flush()
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
		csvField := []string{oldPath, newPath}
		err = csvWriter.Write(csvField)
		if err != nil {
			panic(fmt.Errorf("I couldn't write to the CSV file. Got this error:\n%v", err))
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
