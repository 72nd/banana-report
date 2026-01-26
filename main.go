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
		Engine           string `cli:"--engine, engine to use for PDF generation (typst, fpdf)" default:"typst"`
		DebugCells       bool   `cli:"--debug-cells, enable debug mode for PDF cells"`
		DebugLines       bool   `cli:"--debug-lines, enable debug mode for PDF lines"`
		DebugTempDir     bool   `cli:"--debug-temp-dir, open temp dir in Finder (for typst)"`
		TypstDebug       bool   `cli:"--typst-debug, enable debug mode for typst template"`
		StepEmbedError   bool   `cli:"--step-embed-error, stop on embed error and open file"`
	}
	mcli.Parse(&args)
	dossier, err := DossierFromXML(args.InputPath)
	if err != nil {
		fmt.Println(err)
	}
	if args.Engine == "typst" {
		typst, err := NewTypst(dossier, args.OutputPath, args.TypstDebug)
		if err != nil {
			fmt.Println(err)
		}
		err = typst.Build(args.DebugTempDir)
		if err != nil {
			fmt.Println(err)
		}
		err = typst.Close()
		if err != nil {
			fmt.Println(err)
		}
	} else if args.Engine == "fpdf" {
		pdf := NewPDF(args.CashBasisAccount, args.DebugCells, args.DebugLines, args.StepEmbedError)
		pdf.Build(dossier)
		err = pdf.OutputFileAndClose(args.OutputPath)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Println("invalid engine, available engines: typst, fpdf")
	}
}
