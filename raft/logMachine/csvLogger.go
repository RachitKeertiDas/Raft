package logMachine

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
)

type CsvLogger struct {
	fileName string
}

func (logger *CsvLogger) InitConfig(filename string) {
	logger.fileName = filename

	// Create the file if it doesn't exist already
	f := logger.openFileForAppend()
	defer f.Close()

}

func (logger *CsvLogger) InitLogger() (int, int) {

	fmt.Println("Initializing the Logger.")

	f := logger.openFileForRead()
	defer f.Close()

	csvReader := csv.NewReader(f)
	headers, err := csvReader.Read()
	if err != nil {
		log.Printf("Unexpected Error while reading %s", err)
	}
	err = sanityCheckHeaders(headers)
	if err != nil {
		log.Println("Incorrect headers in CSV File: encountered error", err)
		// Probably should delete the file and try again
	}
	return 0, 0
}

func (logger *CsvLogger) AppendLog(index int, term int, entry string) (int, error) {

	f := logger.openFileForAppend()
	defer f.Close()

	csvWriter := csv.NewWriter(f)

	row := []string{string(term), string(index), entry}
	err := csvWriter.Write(row)
	if err != nil {
		return -1, err
	}

	return index, nil
}

func (logger *CsvLogger) FetchLatestLog() (int, int, string, error) {

	f := logger.openFileForRead()
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return -1, -1, "", err
	}
	// check the first record, it's the header
	header := records[0]
	sanityCheckHeaders(header)

	rawRecords := records[1:]
	// no entries in the file yet
	if len(rawRecords) == 0 {
		return 0, 0, "", nil
	}

	latestRecord := rawRecords[len(records)-1]

	term, _ := strconv.Atoi(latestRecord[0])
	index, _ := strconv.Atoi(latestRecord[1])
	return term, index, latestRecord[2], nil
}

func (logger *CsvLogger) FetchLogWithIndex(index int) (int, string) {

	f := logger.openFileForRead()
	defer f.Close()
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return -1, ""
	}

	for _, record := range records {
		localIndex, _ := strconv.Atoi(record[0])
		if localIndex == index {
			log.Println("found index")
		}
	}

	return 0, ""
}

func (logger *CsvLogger) DeleteEntriesFromIndex(index int) (int, error) {

	f := logger.openFileForAppend()
	defer f.Close()
	return 0, nil
}

func (logger *CsvLogger) openFileForRead() *os.File {
	f, err := os.Open(logger.fileName)
	if err != nil {
		fmt.Println("Error Opening File: %s", err)
		return nil
	}
	return f
}

func (logger *CsvLogger) openFileForAppend() *os.File {
	f, err := os.OpenFile(logger.fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Error Opening File. %s", err)
		return nil
	}
	return f
}

func sanityCheckHeaders(headers []string) error {
	var expectedHeaders = []string{"Term", "Index", "Entry"}

	if slices.Equal(headers, expectedHeaders) {
		return errors.New("Mismatch in headers, expected:" + strings.Join(expectedHeaders, ",") + "found:" + strings.Join(headers, ","))
	}
	return nil

}
