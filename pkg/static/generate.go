// +build ignore

//go:generate go run generate.go

package main

import (
	"log"
	"net/http"
	"path"
	"path/filepath"
	"runtime"

	"github.com/shurcooL/vfsgen"
)

var CRDs http.FileSystem = http.Dir(path.Join(getRepoRoot(), "config/crd/bases"))

func main() {
	rootDir := getRepoRoot()
	log.Printf("rootDir: %s", rootDir)

	err := vfsgen.Generate(CRDs, vfsgen.Options{
		Filename:     path.Join(rootDir, "pkg/static/crds/generated/crds.gogen.go"),
		PackageName:  "generated",
		VariableName: "CRDs",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

// getRepoRoot returns the full path to the root of the repo
func getRepoRoot() string {
	// +nolint
	_, filename, _, _ := runtime.Caller(0)

	dir := filepath.Dir(filename)

	return filepath.Dir(path.Join(dir, ".."))
}
