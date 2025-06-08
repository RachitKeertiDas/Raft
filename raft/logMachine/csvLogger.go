package logMachine

import (
	"encoding/csv"
	"fmt"
	"os"
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
	fmt.Println(csvReader.Read())
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

	return 0, 0, "", nil
}

func (logger *CsvLogger) FetchLogWithIndex(index int) (int, string) {

	f := logger.openFileForRead()
	defer f.Close()

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
