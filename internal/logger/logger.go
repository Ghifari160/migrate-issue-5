package logger

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

const (
	LevelINFO  LogLevel = "INFO"
	LevelWARN  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

const fileDir = "files"
const mainLog = "migrate.log"

const dirPerm = fs.FileMode(0755)
const logPerm = fs.FileMode(0644)

type LogLevel string

// Logger implements a multi-file nested logging system.
// It is concurrency-safe through the use of [sync.Mutex].
// Individual logging files (including the main file) are instances of [logger.LogFile].
type Logger struct {
	dir        string
	main       *LogFile
	filesMutex sync.Mutex
	files      map[string]*LogFile
}

// mkLogDir creates a log directory.
func mkLogDir(path string) error {
	stat, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	} else {
		if err == nil && !stat.IsDir() {
			return errors.New("file " + path + " exists")
		}
	}

	err = os.MkdirAll(filepath.Join(path, fileDir), dirPerm)
	if err != nil {
		return err
	}
	return nil
}

// OpenLogs creates a new Logger, prepares the log directory, and opens the main log file.
func OpenLogs(logDir string) (*Logger, error) {
	l := Logger{}

	logDir, err := filepath.Abs(logDir)
	if err != nil {
		return nil, err
	}

	err = mkLogDir(logDir)
	if err != nil {
		return nil, err
	}
	l.dir = logDir

	f, err := openLogFile(filepath.Join(l.dir, mainLog))
	if err != nil {
		return nil, err
	}
	l.main = f

	l.files = make(map[string]*LogFile)

	return &l, nil
}

// Dir returns the path logging directory.
func (l *Logger) Dir() string {
	l.filesMutex.Lock()
	defer l.filesMutex.Unlock()

	return l.dir
}

// DirAbs returns the absolute path to the logging directory.
func (l *Logger) DirAbs() string {
	abs, err := filepath.Abs(l.dir)
	if err != nil {
		return ""
	}
	return abs
}

// Close closes all open log files including the main log file.
func (l *Logger) Close() error {
	l.filesMutex.Lock()
	defer l.filesMutex.Unlock()

	l.main.Log(LevelINFO, "Closing log files.")

	for _, log := range l.files {
		err := log.Close()
		if err != nil {
			l.main.Log(LevelError, "Error closing log file for "+log.Name()+". Ignoring error.")
		}
	}

	return l.main.Close()
}

// File returns the LogFile for the specified file.
// If the file cannot be found, it is created.
func (l *Logger) File(file string) *LogFile {
	l.filesMutex.Lock()
	defer l.filesMutex.Unlock()

	f, found := l.files[file]
	if !found {
		lF, err := openLogFile(filepath.Join(l.dir, fileDir, file))
		if err != nil {
			format := "Cannot create log file for %s. Logging to main log file instead."
			l.Log(LevelError, fmt.Sprintf(format, file))
			l.Log(LevelError, err.Error())

			lF = l.main
		}

		f = lF
	}

	return f
}

// Write implements io.Writer.
func (l *Logger) Write(entry []byte) (int, error) {
	return l.main.Write(entry)
}

// WriteString implements io.StringWriter.
func (l *Logger) WriteString(entry string) (int, error) {
	return l.main.WriteString(entry)
}

// Log logs an entry at the specified level.
func (l *Logger) Log(level LogLevel, entry string) error {
	return l.main.Log(level, entry)
}
