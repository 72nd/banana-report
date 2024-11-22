package main

import (
	"fmt"

	"github.com/jxskiss/mcli"
)

func main() {
	var args struct {
		InputPath  string `cli:"#R, -i, --input, path to Banana XML file"`
		OutputPath string `cli:"#R, -o, --output, PDF output path"`
	}
	mcli.Parse(&args)
	dossier, err := DossierFromXML(args.InputPath)
	if err != nil {
		fmt.Println(err)
	}
	pdf := PDFFromDossier(*dossier)
	err = pdf.Save(args.OutputPath)
	if err != nil {
		fmt.Println(err)
	}
}
