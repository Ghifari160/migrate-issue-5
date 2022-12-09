package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogFile wraps a text file (i.e. *.log) and provides writing and logging functions.
// It is concurrency-safe through the use of [sync.Mutex].
type LogFile struct {
	m      sync.Mutex
	file   *os.File
	name   string
	closed bool
}

// openLogFile creates a LogFile and opens the underlying file.
func openLogFile(path string) (*LogFile, error) {
	log := LogFile{
		name: path,
	}

	if filepath.Ext(path) != ".log" {
		path += ".log"
	}

	dir := filepath.Dir(path)

	_, err := os.Stat(dir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	err = os.MkdirAll(dir, dirPerm)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, logPerm)
	if err != nil {
		return nil, err
	}
	log.file = file

	return &log, nil
}

// Name returns the name associated with this LogFile.
func (l *LogFile) Name() string {
	l.m.Lock()
	defer l.m.Unlock()

	return l.file.Name()
}

// Close closes the underlying file.
func (l *LogFile) Close() error {
	l.m.Lock()
	defer l.m.Unlock()

	l.closed = true

	return l.file.Close()
}

// Write implements io.Writer.
// Writing to a closed file returns zero length and nil error.
func (l *LogFile) Write(entry []byte) (int, error) {
	l.m.Lock()
	defer l.m.Unlock()

	if l.closed {
		return 0, nil
	}

	return l.file.Write(entry)
}

// WriteString implements io.StringWriter.
// Writing to a closed file returns zero length and nil error.
func (l *LogFile) WriteString(entry string) (int, error) {
	l.m.Lock()
	defer l.m.Unlock()

	if l.closed {
		return 0, nil
	}

	return l.file.WriteString(entry)
}

// Log logs an entry at the specified level.
func (l *LogFile) Log(level LogLevel, entry string) error {
	t := time.Now().Format("2006/01/02T15:04:05.000000")

	_, err := l.WriteString(fmt.Sprintf("%s [%s] %s\n", t, level, entry))
	return err
}
