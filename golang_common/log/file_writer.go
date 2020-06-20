package log

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"time"
)

var pathVariableTable map[byte]func(*time.Time) int

type FileWriter struct {
	logLevelFloor int
	logLevelCeil  int
	filename      string
	pathFmt       string
	file          *os.File
	fileBufWriter *bufio.Writer
	actions       []func(*time.Time) int
	variables     []interface{}
}

func NewFileWriter() *FileWriter {
	return &FileWriter{}
}

func (w *FileWriter) Init() error {
	return w.CreateLogFile()
}

func (w *FileWriter) SetFileName(filename string) {
	w.filename = filename
}

func (w *FileWriter) SetLogLevelFloor(floor int) {
	w.logLevelFloor = floor
}

func (w *FileWriter) SetLogLevelCeil(ceil int) {
	w.logLevelCeil = ceil
}

func (w *FileWriter) SetPathPattern(pattern string) error {
	n := 0
	for _, c := range pattern {
		if c == '%' {
			n++
		}
	}

	if n == 0 {
		w.pathFmt = pattern
		return nil
	}

	w.actions = make([]func(*time.Time) int, 0, n)
	w.variables = make([]interface{}, n, n)
	tmp := []byte(pattern)

	variable := 0
	for _, c := range tmp {
		if variable == 1 {
			act, ok := pathVariableTable[c]
			if !ok {
				return errors.New("Invalid rotate pattern (" + pattern + ")")
			}
			w.actions = append(w.actions, act)
			variable = 0
			continue
		}
		if c == '%' {
			variable = 1
		}
	}

	for i, act := range w.actions {
		now := time.Now()
		w.variables[i] = act(&now)
	}
	//fmt.Printf("%v\n", w.variables)

	w.pathFmt = convertPatternToFmt(tmp)

	return nil
}

func (w *FileWriter) Write(r *Record) error {
	if r.level < w.logLevelFloor || r.level > w.logLevelCeil {
		return nil
	}
	if w.fileBufWriter == nil {
		return errors.New("no opened file")
	}
	if _, err := w.fileBufWriter.WriteString(r.String()); err != nil {
		return err
	}
	return nil
}

func (w *FileWriter) CreateLogFile() error {
	if err := os.MkdirAll(path.Dir(w.filename), 0755); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	if file, err := os.OpenFile(w.filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644); err != nil {
		return err
	} else {
		w.file = file
	}

	if w.fileBufWriter = bufio.NewWriterSize(w.file, 8192); w.fileBufWriter == nil {
		return errors.New("new fileBufWriter failed.")
	}

	return nil
}

func (w *FileWriter) Rotate() error {

	now := time.Now()
	v := 0
	rotate := false
	old_variables := make([]interface{}, len(w.variables))
	copy(old_variables, w.variables)

	for i, act := range w.actions {
		v = act(&now)
		if v != w.variables[i] {
			w.variables[i] = v
			rotate = true
		}
	}
	//fmt.Printf("%v\n", w.variables)

	if rotate == false {
		return nil
	}

	if w.fileBufWriter != nil {
		if err := w.fileBufWriter.Flush(); err != nil {
			return err
		}
	}

	if w.file != nil {
		// 将文件以pattern形式改名并关闭
		filePath := fmt.Sprintf(w.pathFmt, old_variables...)

		if err := os.Rename(w.filename, filePath); err != nil {
			return err
		}

		if err := w.file.Close(); err != nil {
			return err
		}
	}

	return w.CreateLogFile()
}

func (w *FileWriter) Flush() error {
	if w.fileBufWriter != nil {
		return w.fileBufWriter.Flush()
	}
	return nil
}

func getYear(now *time.Time) int {
	return now.Year()
}

func getMonth(now *time.Time) int {
	return int(now.Month())
}

func getDay(now *time.Time) int {
	return now.Day()
}

func getHour(now *time.Time) int {
	return now.Hour()
}

func getMin(now *time.Time) int {
	return now.Minute()
}

func convertPatternToFmt(pattern []byte) string {
	pattern = bytes.Replace(pattern, []byte("%Y"), []byte("%d"), -1)
	pattern = bytes.Replace(pattern, []byte("%M"), []byte("%02d"), -1)
	pattern = bytes.Replace(pattern, []byte("%D"), []byte("%02d"), -1)
	pattern = bytes.Replace(pattern, []byte("%H"), []byte("%02d"), -1)
	pattern = bytes.Replace(pattern, []byte("%m"), []byte("%02d"), -1)
	return string(pattern)
}

func init() {
	pathVariableTable = make(map[byte]func(*time.Time) int, 5)
	pathVariableTable['Y'] = getYear
	pathVariableTable['M'] = getMonth
	pathVariableTable['D'] = getDay
	pathVariableTable['H'] = getHour
	pathVariableTable['m'] = getMin
}
