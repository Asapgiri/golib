package renderer

import (
	"asapgiri/golib/session"
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var artifact_path string = "artifacts/"
var html_path string = "html/"
var base_template_path string = html_path + "base.html"

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

var funcMap = template.FuncMap {
    "inc":      func(i int) int {return i + 1},
    "dec":      func(i int) int {return i - 1},
    "size":     sizeToText,
    "timegt":   func(a time.Time, b time.Time) bool {return b.Compare(a) > 0},
    "timelt":   func(a time.Time, b time.Time) bool {return b.Compare(a) <= 0},
    "now":      time.Now,
    "day":      func() time.Duration {return time.Hour * 24},
    "tformat":  func(a time.Time) string {return a.Local().Format("2006-01-02 15:04:05")},
    "shorten":  func(s string, newLen int) string {return s[:newLen] + ".." + s[len(s)-(newLen/2):]},
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

func Render(session session.Sessioner, w http.ResponseWriter, temp string, dto any) {
    tmp, err := template.ParseFiles(base_template_path)
    if nil != err {
        io.WriteString(w, "Templating error!")
        return
    }

    session.Main = temp
    session.Dto = dto

    var tpl bytes.Buffer
    tmp.Execute(&tpl, session)
    main, err := template.New("Main").Funcs(funcMap).Parse(tpl.String())
    if nil != err {
        io.WriteString(w, "Templating error 2!" + err.Error())
        return
    }

    main.Execute(w, session)
}

func RenderMultiTemplate(session session.Sessioner, w http.ResponseWriter, temp_files []string, dto any) {

    session.Dto = dto

    template_buffer := bytes.Buffer{}
    for _, tf := range(temp_files) {
        fil, _ := ReadArtifact(tf, w.Header())
        temp, err := template.New(tf).Funcs(funcMap).Parse(fil)
        if nil != err {
            io.WriteString(w, "Multi Templating error!" + err.Error())
            return
        }

        session.Main = template_buffer.String()
        template_buffer = bytes.Buffer{}

        temp.Execute(&template_buffer, session)
    }

    session.Main = template_buffer.String()
    main, err := template.ParseFiles(base_template_path)
    if nil != err {
        io.WriteString(w, "Multi Templating error main!" + err.Error())
        return
    }
    main.Execute(w, session)
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
