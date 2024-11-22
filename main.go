package main

import (
	"fmt"

	"github.com/jxskiss/mcli"
)

func main() {
	var args struct {
		InputPath  string `cli:"#R, -i, --input, path to Banana XML file"`
		OutputPath string `cli:"#R, -o, --output, PDF output path"`
		DebugMode  bool   `cli:"--debug, enable debug mode"`
	}
	mcli.Parse(&args)
	dossier, err := DossierFromXML(args.InputPath)
	if err != nil {
		fmt.Println(err)
	}
	pdf := NewPDF(args.DebugMode)
	pdf.Build(dossier)
	err = pdf.OutputFileAndClose(args.OutputPath)
	if err != nil {
		fmt.Println(err)
	}
}
