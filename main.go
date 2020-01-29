package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
)

func main() {
	inputDir := flag.String("target-dir", ".", "Directory where files to rename are found. Defaults to '.'")
	flag.Parse()

	files, err := ioutil.ReadDir(*inputDir)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		matched, err := match(file.Name())
		if err != nil {
			fmt.Printf("`%s` didn't match the pattern I was looking for\n", file)
		}

		if !matched {
			continue
		}
		nameAndExt := strings.Split(file.Name(), ".")
		dateTaken := strings.Split(nameAndExt[0], "_")[0]
		date, err := time.Parse("20060102", dateTaken)
		if err != nil {
			panic(fmt.Errorf("I couldn't parse the date `%s` correctly. Got this error\n %v", dateTaken, err))
		}

		ext := nameAndExt[1]
		newName := fmt.Sprintf("%s--%s.%s", date.Format("January_02_2006"), strings.Split(nameAndExt[0], "_")[1], ext)
		fmt.Println(newName)
	}
}

func match(filename string) (bool, error) {
	isMatch, err := regexp.Match(`^\d{8}_\d+.jpg$`, []byte(filename))
	if err != nil {
		return false, err
	}
	return isMatch, nil
}

func parseDate(date string) (day, month, year string) {
	year = date[0:4]
	month = date[4:6]
	day = date[6:]
	return
}
