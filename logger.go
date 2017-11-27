package logger

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const (
	DEFAULT_RUNTIMECALLER = 1
)

type Logger struct {
	level         int
	err           *log.Logger
	info          *log.Logger
	log           *log.Logger
	curFile       *os.File
	startDate     string
	filename      string
	logDir        string
	runtimeCaller int
	logFilePath   bool // 上报日志的文件位置
	logFunc       bool // 上报日志的方法
	flag          int
	l             sync.Mutex // 锁住curFile文件的修改
}

func initlogger(w io.Writer, flag int) *Logger {
	l := new(Logger)
	l.err = log.New(w, "[ERROR]", flag)
	l.info = log.New(w, "[INFO]", flag)
	l.log = log.New(w, "", 0)
	return l
}

func NewLogger(filename string, logFilePath bool, logFunc bool, isAppend bool) *Logger {
	var err error
	logDir := filepath.Dir(filename)
	result := new(Logger)
	if !result.isExist(filename) {
		// 创建目录

		if !result.isExist(logDir) {
			logAbsDir, err := filepath.Abs(logDir)
			if err != nil {
				fmt.Println(err)
				return result
			} else {
				err = os.MkdirAll(logAbsDir, os.ModePerm)
				if err != nil {
					fmt.Println("创建目录失败:", logAbsDir)
					return result
				}
			}
		}

		//创建文件
		result.curFile, err = os.Create(filename)
		if err != nil {
			fmt.Println(err)
			return result
		}
	} else {
		var rwx int
		if isAppend {
			rwx = os.O_WRONLY | os.O_CREATE | os.O_APPEND
		} else {
			rwx = os.O_WRONLY | os.O_CREATE
		}
		result.curFile, _ = os.OpenFile(filename, rwx, 0666) //打开文件
	}

	result.filename = filename
	result.logDir = logDir
	result.runtimeCaller = DEFAULT_RUNTIMECALLER
	result.logFilePath = logFilePath
	result.logFunc = logFunc
	result.flag = log.Ldate | log.Ltime
	result.startDate = time.Now().Format("2006-01-02")
	result.err = log.New(result.curFile, "[ERROR]", log.Ldate|log.Ltime)
	result.info = log.New(result.curFile, "[INFO]", log.Ldate|log.Ltime)
	result.log = log.New(result.curFile, "", 0)
	return result
}

//String 记录一条纯日志
func (l *Logger) Log(format string, v ...interface{}) {
	l.willSplit()

	l.l.Lock()
	defer l.l.Unlock()
	l.log.Printf(format, v...)
}

//Error 记录一条错误日志
func (l *Logger) Error(format string, v ...interface{}) {
	l.willSplit()

	l.l.Lock()
	defer l.l.Unlock()
	var buf bytes.Buffer
	funcName, file, line, ok := runtime.Caller(l.runtimeCaller)
	if ok {
		if l.logFilePath {
			buf.WriteString(filepath.Base(file))
			buf.WriteString(":")
			buf.WriteString(strconv.Itoa(line))
			buf.WriteString(" ")
		}
		if l.logFunc {
			buf.WriteString(runtime.FuncForPC(funcName).Name())
			buf.WriteString(" ")
		}
		buf.WriteString(format)
		format = buf.String()
	}
	l.err.Printf(format, v...)
}

// 拆分日志文件
func (l *Logger) splitLogFile() {
	l.l.Lock()
	defer l.l.Unlock()

	if l.curFile != nil {
		l.curFile.Close()
		yesterday := time.Now().AddDate(0, 0, -1)
		os.Rename(l.filename, l.filename+"."+yesterday.Format("2006-01-02"))
	}

	logFile, _ := os.Create(l.filename)
	l.curFile = logFile
	l.err = log.New(l.curFile, "[ERROR]", l.flag)
	l.info = log.New(l.curFile, "[INFO]", l.flag)
	l.log = log.New(l.curFile, "", 0)
}

// 检查跨日期，拆分日志文件
func (l *Logger) willSplit() {
	nowDate := time.Now().Format("2006-01-02")
	if l.startDate != nowDate {
		l.startDate = nowDate
		l.splitLogFile()
	}
}

// 检查文件或者文件夹是否存在
func (l *Logger) isExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
