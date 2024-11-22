package main

import (
	_ "embed"
	"path/filepath"

	"github.com/go-pdf/fpdf"
)

//go:embed literata_7pt-regular.ttf
var literataRegular []byte

//go:embed literata_7pt-semi-bold.ttf
var literataSemiBold []byte

type PDF struct {
	*fpdf.Fpdf
	debugMode  bool
	PageWidth  float64
	pageHeight float64
}

func NewPDF(debugMode bool) PDF {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8FontFromBytes("Literata", "", literataRegular)
	pdf.AddUTF8FontFromBytes("Literata", "B", literataSemiBold)

	pageWidth, pageHeight := pdf.GetPageSize()
	return PDF{
		Fpdf:       pdf,
		debugMode:  debugMode,
		PageWidth:  pageWidth,
		pageHeight: pageHeight,
	}
}

func (pdf PDF) Build(dossier *Dossier) {
	for path, doc := range dossier.JournalEntries {
		pdf.addDocument(*dossier, doc, path)
	}
}

func (pdf PDF) addDocument(dossier Dossier, doc Document, path string) {
	pdf.AddPage()
	pdf.Rect(10, 10, pdf.PageWidth-20, pdf.pageHeight-20, "D")
	pdf.addHeader(doc, path)

	pdf.Ln(5)
	pdf.addTableHeader()
	for ident, transactions := range doc {
		pdf.addTableRows(ident, transactions)
	}
}

func (pdf PDF) addHeader(doc Document, path string) {
	pdf.boldFont(12)
	pdf.Ln(1.5)
	pdf.TextCell(pdf.PageWidth-20, 10, filepath.Base(path), "LT")
	pdf.Ln(15)
	pdf.regularFont(10)
	pdf.TextCell(pdf.PageWidth-20, 5, doc.IdentStringList(), "LT")

}

func (pdf PDF) addTableHeader() {
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(30, 7, "Ident", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Datum", "1", 0, "C", true, 0, "")
	pdf.CellFormat(60, 7, "Beschreibung", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Soll", "1", 0, "R", true, 0, "")
	pdf.CellFormat(25, 7, "Haben", "1", 0, "R", true, 0, "")
	pdf.CellFormat(25, 7, "Betrag", "1", 1, "R", true, 0, "")
}

func (pdf PDF) addTableRows(ident string, transactions []Transaction) {
	pdf.SetFont("Arial", "", 10)

	// Merged cell for "Ident"
	pdf.CellFormat(30, 7, ident, "L", 0, "C", false, 0, "")

	// First row of the group
	first := true
	for _, tx := range transactions {
		if !first {
			// Empty Ident cell for subsequent rows
			pdf.CellFormat(30, 7, "", "", 0, "C", false, 0, "")
		}
		// Add transaction details
		pdf.CellFormat(30, 7, tx.Date, "", 0, "C", false, 0, "")
		pdf.CellFormat(60, 7, tx.Description, "", 0, "L", false, 0, "")
		pdf.CellFormat(25, 7, tx.AccountDebit, "", 0, "R", false, 0, "")
		pdf.CellFormat(25, 7, tx.AccountCredit, "", 0, "R", false, 0, "")
		pdf.CellFormat(25, 7, tx.Amount, "R", 1, "R", false, 0, "")
		first = false
	}

	// Draw dotted separator
	pdf.SetDrawColor(150, 150, 150)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.SetDrawColor(0, 0, 0)
}

func (pdf PDF) Lol(document Document) {

	pdf.regularFont(7)
	pageWidth, _ := pdf.GetPageSize()

	// Table settings
	leftMargin := 10.0
	top := 50.0
	rowHeight := 4.5
	tableWidth := pageWidth - 2*leftMargin
	col1Width := 50.0
	col2Width := (tableWidth - col1Width) / 2
	col3Width := col2Width

	// Example data
	mergedText := document.IdentStringList() + "\n\n "
	rows := [][]string{
		{"", "Data 1", "Data 2"},
		{"", "Data 3", "Data 4"},
		{"", "Data 5", "Data 6"},
	}

	// Draw merged column 1
	pdf.SetXY(leftMargin, top)
	pdf.MultiCell(col1Width, rowHeight, mergedText, "", "L", false)

	// Draw other columns for each row
	currentY := top
	for _, row := range rows {
		// Top border (dashed)
		pdf.SetDashPattern([]float64{1, 1}, 0)
		pdf.Line(leftMargin+col1Width, currentY, leftMargin+tableWidth, currentY)

		// Reset to solid for the bottom border
		pdf.SetDashPattern([]float64{}, 0)
		bottomY := currentY + rowHeight
		pdf.Line(leftMargin+col1Width, bottomY, leftMargin+tableWidth, bottomY)

		// Draw text in column 2 and column 3
		currentX := leftMargin + col1Width
		for i, cell := range row[1:] {
			pdf.SetXY(currentX, currentY)
			width := col2Width
			if i == 1 {
				width = col3Width
			}
			pdf.CellFormat(width, rowHeight, cell, "0", 0, "C", false, 0, "")
			currentX += width
		}

		// Move to next row
		currentY += rowHeight
	}
}

func (pdf PDF) TextCell(w, h float64, txtStr string, alignStr string) {
	drawR, drawG, drawB := pdf.GetDrawColor()
	borderStr := ""
	if pdf.debugMode {
		borderStr = "1"
		pdf.SetDrawColor(0xDA, 0x6A, 0x35)
	}
	pdf.CellFormat(w, h, txtStr, borderStr, 0, alignStr, false, 0, "")
	if pdf.debugMode {
		pdf.SetDrawColor(drawR, drawG, drawB)
	}
}

func (pdf PDF) regularFont(size float64) {
	pdf.SetFont("Literata", "", size)
}

func (pdf PDF) boldFont(size float64) {
	pdf.SetFont("Literata", "B", size)
}
