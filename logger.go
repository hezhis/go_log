package logger

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	DebugLevel = iota + 1 // Debug级别
	InfoLevel             // Info级别
	WarnLevel             // Warn级别
	ErrorLevel            // Error级别
	FatalLevel            // Fatal级别
)

const (
	red = string(byte(27)) + "[" + string(byte(31)) + "m"
	//green   = "\27[32m"
	//yellow  = "\27[33m"
	//blue    = "\27[34m"
	//magenta = "\27[35m"
	//cyan    = "\27[36m"
	//white   = "\27[37m"
	//clear   = "\27[0m"
)
const (
	fileMode = 0777
)

var (
	_logger   *log.Logger
	_logFile  *os.File //日志文件
	_logLevel = DebugLevel
	_logFlag  = log.LstdFlags | log.Lshortfile
	_logName  string //日志名称
	_logPath  string
	_stdout   bool
)

func Setup(path, name string) {
	if "" == path {
		_logPath = curDir()
	} else {
		_logPath = path
	}

	_, err := os.Stat(_logPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(_logPath, fileMode)
	}

	if nil != err {
		log.Fatal(fmt.Sprintf("logger setup error! %v", err))
		return
	}
	_logName = name

	_logger = log.New(nil, "", _logFlag)

	UpdateLogFile()
}

func Close() {
	if nil != _logFile {
		_logFile.Close()
	}
	_logFile = nil
	_logger = nil
}

func SetLevel(level int) {
	if level > FatalLevel || level < DebugLevel {
		_logLevel = DebugLevel
	} else {
		_logLevel = level
	}
}

func UpdateLogFile() {
	if "" != _logName {
		t := time.Now()
		lfName := fmt.Sprintf("%s/%s%04d%02d%02d.log", _logPath, _logName, t.Year(), t.Month(), t.Day())
		f, err := os.OpenFile(lfName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, fileMode)
		if _logFile == nil {
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		} else {
			_logFile.Close()
		}
		if err == nil {
			//设置新的文件
			_logFile = f
			_logger.SetOutput(_logFile)
		}
		_stdout = false
	} else {
		_logger = log.New(os.Stdout, "", _logFlag)
		_stdout = true
	}
}

func Debug(format string, a ...interface{}) {
	output(DebugLevel, "<DEBUG> ", "\u001B[32m<DEBUG> \u001B[0m", format, a...)
}

func Info(format string, a ...interface{}) {
	output(InfoLevel, "<INFO> ", "\u001B[32m<INFO> \u001B[0m", format, a...)
}

func Warn(format string, a ...interface{}) {
	output(WarnLevel, "<WARN> ", "\u001B[33m<WARN> \u001B[0m", format, a...)
}

func Error(format string, a ...interface{}) {
	output(ErrorLevel, "<ERROR> ", "\u001B[31m<ERROR> \u001B[0m", format, a...)
}

func Fatal(format string, a ...interface{}) {
	output(FatalLevel, "<FATAL> ", "\u001B[31m<FATAL> \u001B[0m", format, a...)
}

func output(level int, prefix, colorPrefix, format string, a ...interface{}) {
	if _logLevel > level {
		return
	}
	if nil == _logger {
		return
	}

	content := fmt.Sprintf(format, a...)

	var builder strings.Builder
	builder.WriteString(prefix)
	builder.WriteString(content)
	_logger.Output(3, builder.String())

	if !_stdout && runtime.GOOS == "windows" {
		log.Println(colorPrefix + content)
	}

	if FatalLevel == level {
		buf := make([]byte, 4096)
		l := runtime.Stack(buf, false)
		_logger.Output(3, string(buf[:l]))

		tf := time.Now()
		ioutil.WriteFile(fmt.Sprintf("%s/core-%s.%02d%02d-%02d%02d%02d.panic", curDir(), _logName,
			tf.Month(), tf.Day(), tf.Hour(), tf.Minute(), tf.Second()), []byte(builder.String()+"\n"+string(buf[:l])), fileMode)

		os.Exit(1)
	}
}

//获取当前目录
func curDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return ""
	}
	return strings.Replace(dir, "\\", "/", -1) + "/"
}

func detailInfo() string {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[1])
	file, line := f.FileLine(pc[1])
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			file = file[i+1:]
			break
		}
	}
	funcName := f.Name()
	for i := len(funcName) - 1; i > 0; i-- {
		if funcName[i] == '.' {
			funcName = funcName[i+1:]
			break
		}
	}
	return fmt.Sprintf("%s [%s:%d %s]", time.Now().Format("01-02 15:04:05"), file, line, funcName)
}

func init() {
	Setup("", "")
}
