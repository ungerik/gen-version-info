//go:generate go run gen-version-info.go
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ungerik/go-dry"
)

const ISO8601 = "2006-01-02 15:04:05 -0700"

var (
	lang     = flag.String("lang", "go", "Programming language")
	basename = flag.String("file", "version", "Base name of the generated file, a language specific extension will be added")
)

func cmd(name string, args ...string) string {
	res, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("%s %s\n", err, res))
	}
	return strings.TrimSpace(string(res))
}

func findGit() bool {
	dir, err := os.Getwd()
	for !dry.FileIsDir(path.Join(dir, ".git")) {
		dir, err = filepath.Abs(path.Join(dir, ".."))
		if dir == "" || dir == "/" || err != nil {
			return false
		}
	}
	return true
}

func main() {
	flag.Parse()

	var buf bytes.Buffer

	if findGit() {

		version := cmd("git", "describe", "--tags", "--always")
		bt := time.Now().UTC()
		vt, err := time.Parse(ISO8601, cmd("git", "show", "-s", "--format=%ci"))
		if err != nil {
			panic(err)
		}
		vt = vt.UTC()

		fprintVersion := func(writer io.Writer) {
			fmt.Fprintf(writer, "package main\n\nimport \"time\"\n\nconst (\n")
			fmt.Fprintf(writer, "\tVERSION                = \"%s\"\n", version)
			fmt.Fprintf(writer, "\tVERSION_CONTROL_SYSTEM = \"git\"\n")
			fmt.Fprintf(writer, ")\nvar (\n")
			fmt.Fprintf(writer, "\tVERSION_TIME        = time.Date(%d, %d, %d, %d, %d, %d, 0, time.UTC)\n", vt.Year(), vt.Month(), vt.Day(), vt.Hour(), vt.Minute(), vt.Second())
			fmt.Fprintf(writer, "\tVERSION_BUILD_TIME  = time.Date(%d, %d, %d, %d, %d, %d, 0, time.UTC)\n", bt.Year(), bt.Month(), bt.Day(), bt.Hour(), bt.Minute(), bt.Second())
			fmt.Fprintf(writer, ")\n")
		}

		fprintVersion(&buf)

		filename := *basename + ".go"

		err = dry.FileSetBytes(filename, buf.Bytes())
		if err != nil {
			panic(err)
		} else {
			fmt.Println("Created file", filename)
			fprintVersion(os.Stdout)
		}

	} else if dry.FileIsDir(".svn") {
	} else {
		panic("Need .git or .svn directory")
	}

}
