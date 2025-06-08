package core

// Defines the logger Interface to store log entries.
type LogMachineHandler interface {

	// Init state.
	// When a server starts up, we want to init the logMachine.
	// In case it has started up after a crash, but we can still
	// access the logs, Init() should be able to read current logs
	// and return the last known `term` and `logIndex`
	// logIndex - refers to the index of the last entry
	InitLogger() (int, int)

	// append a new entry to the log machine
	// Takes in the term, supposed index and the string
	// returns the index of the new entry
	AppendLog(int, int, string) (int, error)

	// Fetch the Latest Log entry present
	// returns the term, index, and the entry itself
	FetchLatestLog() (int, int, string, error)

	// given an index, fetches the log entry present at that index
	// if there is no entry present at that index, return nil
	FetchLogWithIndex(int) (int, string)
	// TODO: Should there be a FetchLogsFromIndex?

	// Delete all the Entries starting at a particular index
	// Useful when the state machines has to overwrite entries
	// beginning at a particular index.
	DeleteEntriesFromIndex(int) (int, error)
}
