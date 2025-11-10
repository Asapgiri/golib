package renderer

import (
	"github.com/asapgiri/golib/session"
	"github.com/asapgiri/golib/logger"
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var artifact_path string = "artifacts/"
var html_path string = "html/"
var base_template_path string = html_path + "base.html"
var template_files_path string = html_path + "templates/"

var log = logger.Logger {
    Color: logger.Colors.Brown_Orange,
    Pretext: "renderer",
}

var file_types = map[string]string {
    "html": "text",
    "css":  "text",
}

func sizeToText(size int) string {
    const kbDiv = 1024.0

    mb := float64(size) / kbDiv / kbDiv
    gb := mb / kbDiv

    if gb >= 1.0 {
        return strconv.FormatFloat(gb, 'f', 2, 64) + " GB"
    } else {
        return strconv.FormatFloat(mb, 'f', 2, 64) + " MB"
    }
}

func seq(start, end int) []int {
    s := make([]int, 0, end-start)
    for i := start; i < end; i++ {
        s = append(s, i)
    }
    return s
}

func Subset(a []string, b []string) bool {
    for _, s := range(a) {
        if slices.Contains(b, s) {
            return true
        }
    }

    return false
}

var funcMap = template.FuncMap {
    "inc":      func(i int) int {return i + 1},
    "dec":      func(i int) int {return i - 1},
    "seq":      seq,
    "size":     sizeToText,
    "timegt":   func(a time.Time, b time.Time) bool {return b.Compare(a) > 0},
    "timelt":   func(a time.Time, b time.Time) bool {return b.Compare(a) <= 0},
    "now":      time.Now,
    "day":      func() time.Duration {return time.Hour * 24},
    "tformat":  func(a time.Time) string {return a.Local().Format("2006-01-02 15:04:05")},
    "shorten":  func(s string, newLen int) string {return s[:newLen] + ".." + s[len(s)-(newLen/2):]},
    "scon":     func(arr []string, x string) bool {return slices.Contains(arr, x)},
    "subset":   Subset,
}

func ReadArtifact(path string, header http.Header) (string, string) {
    var dir_path string

    ex, err := os.Executable()
    if nil != err {
        panic(err)
    }

    parts := strings.Split(path, ".")
    file_type := parts[len(parts)-1]
    if "html" == file_type {
        dir_path = filepath.Dir(ex) + "/" + html_path
    } else {
        dir_path = filepath.Dir(ex) + "/" + artifact_path
    }

    file_read, err := os.ReadFile(dir_path + path)
    if nil != err {
        not_found, _ := os.ReadFile(filepath.Dir(ex) + "/" + html_path + "not_found.html")
        return string(not_found), "text"
    }

    if nil != header {
        _, file_ok := file_types[file_type]
        if file_ok {
            header.Set("Content-Type", file_types[file_type] + "/" + file_type)
        }
    }

    return string(file_read), file_type
}

func SaveArtifact(path string, file multipart.File) error {
    ex, err := os.Executable()
    if nil != err {
        panic(err)
    }

    dir_path := filepath.Dir(ex) + "/" + artifact_path
    dstPath := filepath.Join(dir_path, path)
    dst, err := os.Create(dstPath)
    if nil != err {
        return err
    }
    defer dst.Close()

    _, err = io.Copy(dst, file)
    if nil != err {
        return err
    }

    return nil
}

func get_template_files() []string {
    entries, err := os.ReadDir(template_files_path)
    if nil != err {
        return []string{}
    }

    var files []string
    for _, e := range(entries) {
        if !e.IsDir() {
            files = append(files, filepath.Join(template_files_path, e.Name()))
        }
    }

    return files
}

func Render(session session.Sessioner, wr io.Writer, temp string, dto any) error {
    tmp, err := template.ParseFiles(base_template_path)
    if nil != err {
        log.Println(err)
        io.WriteString(wr, "Templating error!")
        return err
    }

    session.Main = temp
    session.Dto = dto

    var tpl bytes.Buffer
    tmp.Execute(&tpl, session)

    main, err := template.New("Main").Funcs(funcMap).ParseFiles(get_template_files()...)
    if nil != err {
        log.Println(err)
        io.WriteString(wr, "Templating error 2!" + err.Error())
        return err
    }
    main, err = main.Parse(tpl.String())
    if nil != err {
        log.Println(err)
        io.WriteString(wr, "Templating error 3!" + err.Error())
        return err
    }

    main.Execute(wr, session)
    return nil
}

func RenderMultiTemplate(session session.Sessioner, wr io.Writer, temp_files []string, dto any) {

    session.Dto = dto

    template_buffer := bytes.Buffer{}
    for _, tf := range(temp_files) {
        fil, _ := ReadArtifact(tf, nil)
        temp, err := template.New(tf).Funcs(funcMap).Parse(fil)
        if nil != err {
            io.WriteString(wr, "Multi Templating error!" + err.Error())
            return
        }

        session.Main = template_buffer.String()
        template_buffer = bytes.Buffer{}

        temp.Execute(&template_buffer, session)
    }

    session.Main = template_buffer.String()
    main, err := template.ParseFiles(base_template_path)
    if nil != err {
        io.WriteString(wr, "Multi Templating error main!" + err.Error())
        return
    }
    main.Execute(wr, session)
}

// Prerender does not support session if you don't pass it...
func PreRender(temp string, dto any) string {
    var tpl bytes.Buffer

    tmp, err := template.New("Dto").Funcs(funcMap).Parse(temp)
    if nil != err {
        return err.Error()
    }
    tmp.Execute(&tpl, dto)

    return tpl.String()
}
