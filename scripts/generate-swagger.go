// +build tools

package main

import (
	"log"
	"flag"
	"github.com/jessevdk/go-flags"
	"github.com/go-swagger/go-swagger/cmd/swagger/commands/generate"
)

//TODO make it work
//TODO vendor
//TODO replace like make

func main() {
	var destination string

	flag.StringVar(&destination, "o", "templates/swagger/v1_json.tmpl", "output file")
	flag.Parse()

	var spec = generate.SpecFile{
		BasePath: "./",
		Output: flags.Filename(destination),
	}

	err := spec.Execute([]string{});
	if err != nil {
		log.Fatalf("Failed to generate swagger spec. %s", err)
	}
}
