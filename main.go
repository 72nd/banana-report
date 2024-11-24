package main

import (
	"fmt"

	"github.com/jxskiss/mcli"
)

func main() {
	var args struct {
		InputPath  string `cli:"#R, -i, --input, path to Banana XML file"`
		OutputPath string `cli:"#R, -o, --output, PDF output path"`
		DebugCells bool   `cli:"--debug-cells, enable debug mode for PDF cells"`
		DebugLInes bool   `cli:"--debug-lines, enable debug mode for PDF lines"`
	}
	mcli.Parse(&args)
	dossier, err := DossierFromXML(args.InputPath)
	if err != nil {
		fmt.Println(err)
	}
	pdf := NewPDF(args.DebugCells, args.DebugLInes)
	pdf.Build(dossier)
	err = pdf.OutputFileAndClose(args.OutputPath)
	if err != nil {
		fmt.Println(err)
	}
}
