package main

import (
	"fmt"

	"github.com/jxskiss/mcli"
)

func main() {
	var args struct {
		InputPath        string `cli:"#R, -i, --input, path to Banana XML file"`
		OutputPath       string `cli:"#R, -o, --output, PDF output path"`
		CashBasisAccount bool   `cli:"--cash-basis, switch to cash basis account (EÜR)"`
		DebugCells       bool   `cli:"--debug-cells, enable debug mode for PDF cells"`
		DebugLines       bool   `cli:"--debug-lines, enable debug mode for PDF lines"`
		StepEmbedError   bool   `cli:"--step-embed-error, stop on embed error and open file"`
	}
	mcli.Parse(&args)
	dossier, err := DossierFromXML(args.InputPath)
	if err != nil {
		fmt.Println(err)
	}
	pdf := NewPDF(args.CashBasisAccount, args.DebugCells, args.DebugLines, args.StepEmbedError)
	pdf.Build(dossier)
	err = pdf.OutputFileAndClose(args.OutputPath)
	if err != nil {
		fmt.Println(err)
	}
}
