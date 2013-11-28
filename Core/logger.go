package Core

import (
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

type LoggerInterface interface {
	Log(v...interface {})
	Error(v...interface {})
}

type Logger struct {
	dir       string
	accessLog []string
	errorLog  []string
	clockTime int
	debug     bool
	sync.RWMutex
}

// log something
func (this *Logger) Log(v ...interface{}) {
	log.Println(v...)
	str := fmt.Sprintln(v...)
	this.accessLog = append(this.accessLog, str)
}

// log something error
func (this *Logger) Error(v ...interface{}) {
	log.Println(append([]interface{}{"[E]"}, v...))
	str := fmt.Sprintln(v...)
	this.errorLog = append(this.errorLog, str)
}

// get current log file name
func (this *Logger) GetLogFile() string {
	fileName := "access_" + time.Now().Format("2006_01_02")
	fileName += ".log"
	return path.Join(this.dir, fileName)
}

// get error file name
func (this *Logger) GetErrorLogFile() string {
	return path.Join(this.dir, "error.log")
}

// flush logging messages in memory
func (this *Logger) Flush() {
	this.RLock()
	defer this.RUnlock()
	if len(this.accessLog) > 0 {
		fs, e := os.OpenFile(this.GetLogFile(), os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0660)
		if e != nil {
			panic(e)
		}
		for _, log := range this.accessLog {
			fs.Write([]byte(log))
		}
		this.accessLog = []string{}
		defer fs.Close()
	}
	if len(this.errorLog) > 0 {
		fs, e := os.OpenFile(this.GetErrorLogFile(), os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0660)
		if e != nil {
			panic(e)
		}
		for _, log := range this.errorLog {
			fs.Write([]byte(log))
		}
		this.errorLog = []string{}
		defer fs.Close()
	}
}

// write logs in file by clock time
func (this *Logger) clock() {
	logsLength := len(this.accessLog) + len(this.errorLog)
	if this.debug {
		log.Println("logger flush @", logsLength, "logs")
	}
	this.Flush()
	// redo after clock time
	time.AfterFunc(time.Duration(this.clockTime) * time.Second, func() {
			this.clock()
		})
}

// create new logger
func NewLogger(dir string, clockTime int) *Logger {
	logger := &Logger{}
	logger.dir = dir
	logger.clockTime = clockTime
	logger.accessLog = make([]string, 0)
	logger.errorLog = make([]string, 0)
	logger.clock()
	return logger
}
