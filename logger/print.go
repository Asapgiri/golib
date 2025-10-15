package logger

import (
	"fmt"
	"runtime"
)

type colors struct {
    Black        string
    Red          string
    Green        string
    Brown_Orange string
    Blue         string
    Purple       string
    Cyan         string
    Light_Gray   string

    Dark_Gray    string
    Light_Red    string
    Light_Green  string
    Yellow       string
    Light_Blue   string
    Light_Purple string
    Light_Cyan   string
    White        string

    Default      string
}

var Colors = colors {
    Black        : "\033[0;30m",
    Red          : "\033[0;31m",
    Green        : "\033[0;32m",
    Brown_Orange : "\033[0;33m",
    Blue         : "\033[0;34m",
    Purple       : "\033[0;35m",
    Cyan         : "\033[0;36m",
    Light_Gray   : "\033[0;37m",

    Dark_Gray    : "\033[1;30m",
    Light_Red    : "\033[1;31m",
    Light_Green  : "\033[1;32m",
    Yellow       : "\033[1;33m",
    Light_Blue   : "\033[1;34m",
    Light_Purple : "\033[1;35m",
    Light_Cyan   : "\033[1;36m",
    White        : "\033[1;37m",

    Default      : "\033[0m",
}

type Logger struct {
    Color string
    Pretext string
}

func (l Logger) begin() {
    fmt.Print(l.Color)
    if "" != l.Pretext {
        _, _, line, _ := runtime.Caller(2)
        fmt.Printf("[%s:%d]: ", l.Pretext, line)
    }
}

func (l Logger) end(nl ...bool) {
    if 0 == len(nl) || nl[0] {
        fmt.Println(Colors.Default)
    } else {
        fmt.Print(Colors.Default)
    }
}

func (l Logger) Println(params ...any) {
    l.begin()
    fmt.Print(params ...)
    l.end()
}

func (l Logger) Printf(msg string, params ...any) {
    l.begin()
    fmt.Printf(msg, params ...)
    l.end(false)
}
