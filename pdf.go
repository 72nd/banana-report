package main

import (
	"fmt"

	"github.com/go-pdf/fpdf"
)

type PDF struct {
	doc *fpdf.Fpdf
}

func PDFFromDossier(dossier Dossier) PDF {
	fmt.Println(dossier.AccountingFilePath)
	doc := fpdf.New("P", "mm", "A4", "")

	return PDF{
		doc: doc,
	}
}

func (p PDF) Save(path string) error {
	return p.doc.OutputFileAndClose(path)
}
