package finderhttp

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"
)

type Logger interface {
	Debug(args ...any)
	Info(args ...any)
	Error(args ...any)
}

type logkv struct {
	time  string
	args  []interface{}
	level string
}

const (
	LogDebug = iota
	LogInfo
	LogError
)

type fileLog struct {
	buffer chan *logkv
	file   *bufio.Writer
	level  int
}

func NewLogger(file string, level int) (Logger, error) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	f.Seek(0, io.SeekEnd)
	l := &fileLog{buffer: make(chan *logkv, 10000), level: level, file: bufio.NewWriterSize(f, 1024*128)}
	go l.start()
	return l, nil
}

func (l *fileLog) start() {
	for {
		// 将channel 中的所有数据全部写入文件，再flush
	here:
		for {
			select {
			case kv := <-l.buffer:
				l.writeToFile(kv)
			default:
				l.file.Flush()
				break here
			}
		}

		select {
		case kv := <-l.buffer:
			l.writeToFile(kv)
		}
	}
}

func (l *fileLog) writeToFile(kv *logkv) {
	l.file.WriteString(kv.level)
	l.file.WriteString(kv.time)
	l.file.WriteString(" ")
	fmt.Fprintln(l.file, kv.args...)
}

func (l *fileLog) Debug(args ...any) {
	if l.level > LogDebug {
		return
	}
	l.log("[DEBUG]", args...)
}

func (l *fileLog) Info(args ...any) {
	if l.level > LogInfo {
		return
	}
	l.log("[INFO]", args...)
}

func (l *fileLog) Error(args ...any) {
	if l.level > LogError {
		return
	}
	l.log("[ERR]", args...)
}

func (l *fileLog) log(lv string, args ...any) {
	select {
	case l.buffer <- &logkv{time: time.Now().Format(time.RFC3339), args: args, level: lv}:
	default:

	}
}
