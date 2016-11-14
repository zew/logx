package logx

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

const leadingDirsInPrefix = 2 // how many directories to display in prefix

const ColWidth = 6 // default column width

// os.Stderr is important for app engine
var l = log.New(os.Stderr, "", log.Lshortfile)

func init() {
	l.SetFlags(log.Ltime | log.Lshortfile)
	l.SetFlags(log.Lshortfile)
	l.SetFlags(0)
}

// Returns the the source file.
// Good to read file inside libaries,
// completely independent of working dir
// or application dir.
func PathToSourceFile(levelsUp ...int) string {
	lvlUp := 1
	if len(levelsUp) > 0 {
		lvlUp = 1 + levelsUp[0]
	}
	_, srcFile, _, ok := runtime.Caller(lvlUp)
	if !ok {
		Fatalf("runtime caller not found")
	}
	p := path.Dir(srcFile)
	return p
}

const minCX = 2 // minimum spaces towards next columns

// Should go to package util strings -
// but that causes import cycles
func Columnify(arg string, minWidth, colWidth int) string {
	if len(arg) < minWidth {
		padd := minWidth + minCX - len(arg)
		arg = fmt.Sprintf("%s%s", arg, strings.Repeat(" ", padd))
	} else {
		largestFraction := (len(arg)+minCX)/colWidth + 1
		padd := largestFraction*colWidth - len(arg) // columns of colWidth chars
		arg = fmt.Sprintf("%s%s", arg, strings.Repeat(" ", padd))
	}
	return arg
}

// We dont want 20 leading directories of a source file.
// But the filename alone is not enough.
// "main.go" does not help.
func leadDirsBeforeSourceFile(path string, dirsBeforeSourceFile int) string {
	rump := path // init
	dirs := make([]string, 0, dirsBeforeSourceFile)
	for i := 0; i < dirsBeforeSourceFile; i++ {
		rump = filepath.Dir(rump)
		dir := filepath.Base(rump)
		dirs = append([]string{dir}, dirs...)
	}
	lastDirs := filepath.Join(dirs...)
	lastDirs = filepath.Join(lastDirs, filepath.Base(path))
	return lastDirs
}

// Source code location
// x steps up the call stack
func stackLine(levelUp, dirsBeforeSourceFile int) (int, string) {
	_, file, line, _ := runtime.Caller(levelUp + 1) // plus one for myself-func
	return line, leadDirsBeforeSourceFile(file, dirsBeforeSourceFile)
}

// Show short source path as prefix for each log message
func sourceLocationPrefix() string {
	// line, file := stackLine(2, leadingDirsInPrefix) // me and call; but util.CheckErr needs 3
	line, file := stackLine(sl.lvl, leadingDirsInPrefix)
	prfx := fmt.Sprintf("%s:%d", file, line)
	prfx = Columnify(prfx, 8, ColWidth)
	return prfx
}

// For giving it to a tracer
func Get() *log.Logger {
	return l
}

// Enable default out writer
func LogToStdOut() {
	l.SetOutput(os.Stdout)
}

// Enable specific out writers (i.e. for multi writer)
func LogTo(w io.Writer) {
	l.SetOutput(w)
}

func Disable() {
	l.SetOutput(ioutil.Discard)
}

func Fatalf(format string, v ...interface{}) {
	payload := fmt.Sprintf(format, v...)
	payload = fmt.Sprintf("%s%s%s\n", sourceLocationPrefix(), payload, SPrintStackTrace(2, 5, 2))
	l.Print(payload)
	panic(payload) // Hand the panic up to outer panic handler
}

// We need this to prevent % escaping
func Fatal(v ...interface{}) {
	payload := fmt.Sprint(v...) // only difference to Fatalf
	payload = fmt.Sprintf("%s%s%s\n", sourceLocationPrefix(), payload, SPrintStackTrace(2, 5, 2))
	l.Print(payload)
	panic(payload) // Hand the panic up to outer panic handler
}

func Println(v ...interface{}) {
	asSlice := []interface{}{sourceLocationPrefix()}
	v = append(asSlice, v...)
	l.Println(v...)
}
func Print(v ...interface{}) {
	asSlice := []interface{}{sourceLocationPrefix()}
	v = append(asSlice, v...)
	l.Print(v...)
}

func Printf(format string, args ...interface{}) {
	payload := fmt.Sprintf(format, args...)
	payload = strings.TrimRight(payload, "\n")
	payload = Columnify(payload, 56, 4)
	payload = fmt.Sprintf("%s%s", sourceLocationPrefix(), payload)
	l.Print(payload)
}

//
func PrintStackTrace(args ...int) {
	str := SPrintStackTrace(args...)
	Printf(str)
}

func SPrintStackTrace(args ...int) string {
	lines := stackTrace(args...)
	str := strings.Join(lines, "\n\t")
	str = fmt.Sprintf("\n\t%v", str)
	return str
}

// First  arg => level init
// Second arg => levels up
// Third  arg => dirs of before source file
func stackTrace(args ...int) []string {

	var (
		lvlInit              = 2 // One for this func, one since direct caller is already logged in prefix
		lvlsUp               = 4
		dirsBeforeSourceFile = 2 // How many dirs are shown before the source file.
	)

	if len(args) > 0 {
		lvlInit += args[0]
	}
	if len(args) > 1 {
		lvlsUp = args[1]
	}
	if len(args) > 2 {
		dirsBeforeSourceFile = args[2]
	}

	lines := make([]string, lvlsUp)
	for i := 0; i < lvlsUp; i++ {

		_, file, line, _ := runtime.Caller(i + lvlInit)
		if line == 0 && file == "." {
			break
		}
		file = leadDirsBeforeSourceFile(file, dirsBeforeSourceFile)

		lines[i] = fmt.Sprintf("%s:%d", file, line)
		lines[i] = Columnify(lines[i], 12, 12)
	}
	return lines
}

// All log prints also appear in the http response
func LogToResponseBody(w http.ResponseWriter) {
	// file, err := os.Create("./01.log")
	file, err := os.OpenFile("./01.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic("could not open log file")
	}

	if false {
		w.Write([]byte("init"))
	}

	bodyWtr := io.Writer(w)
	multi := io.MultiWriter(file, os.Stdout, bodyWtr)
	multi = io.MultiWriter(file, os.Stdout)
	LogTo(multi)
}
