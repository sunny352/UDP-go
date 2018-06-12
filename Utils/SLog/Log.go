package SLog

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
)

var std = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)

func D(v ...interface{}) {
	std.Output(2, fmt.Sprintln(v...))
}
func I(v ...interface{}) {
	std.Output(2, fmt.Sprintln(v...))
}
func W(v ...interface{}) {
	std.Output(2, fmt.Sprintln(v...)+string(debug.Stack()))
}
func E(v ...interface{}) {
	std.Output(2, fmt.Sprintln(v...)+string(debug.Stack()))
}
