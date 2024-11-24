package main

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/boombuler/barcode/qr"
	"github.com/go-pdf/fpdf"
	"github.com/go-pdf/fpdf/contrib/barcode"
)

//go:embed literata_7pt-regular.ttf
var literataRegular []byte

//go:embed literata_7pt-italic.ttf
var literataItalic []byte

//go:embed literata_7pt-semi-bold.ttf
var literataSemiBold []byte

type PDF struct {
	*fpdf.Fpdf
	debugMode    bool
	PageWidth    float64
	PageHeight   float64
	FontFamily   string
	LeftMargin   float64
	TopMargin    float64
	RightMargin  float64
	BottomMargin float64
}

func NewPDF(debugMode bool) PDF {
	fontName := "Literata"
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8FontFromBytes(fontName, "", literataRegular)
	pdf.AddUTF8FontFromBytes(fontName, "I", literataItalic)
	pdf.AddUTF8FontFromBytes(fontName, "B", literataSemiBold)

	pageWidth, pageHeight := pdf.GetPageSize()
	lm, tm, rm, bm := pdf.GetMargins()
	return PDF{
		Fpdf:         pdf,
		debugMode:    debugMode,
		PageWidth:    pageWidth,
		PageHeight:   pageHeight,
		FontFamily:   fontName,
		LeftMargin:   lm,
		TopMargin:    tm,
		RightMargin:  rm,
		BottomMargin: bm,
	}
}

func (pdf PDF) Build(dossier *Dossier) {
	for _, doc := range dossier.JournalEntries {
		// FOR DEBUGGING
		title := strings.TrimSuffix(filepath.Base(doc.Path), filepath.Ext(doc.Path))
		if !strings.HasPrefix(title, "hetzner_2023-04-01_") {
		}
		pdf.addDocument(*dossier, doc)
	}
}

func (pdf PDF) addDocument(dossier Dossier, doc Document) {
	pdf.AddPage()
	pdf.Rect(10, 10, pdf.PageWidth-20, pdf.PageHeight-20, "D")
	pdf.addHeader(doc, doc.Path)

	pdf.addTableHeader(4.5)
	pdf.addTableRows(doc.Transactions, 4.5)
}

func (pdf PDF) addHeader(doc Document, path string) {
	title := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	description1 := fmt.Sprintf("%s — ", path)
	description2 := doc.IdentStringList()
	description := fmt.Sprintf("%s%s", description1, description2)
	qrBlockDimensions := 15.
	textBlockWidth := pdf.PageWidth - pdf.LeftMargin - pdf.RightMargin - qrBlockDimensions
	qrBlockX := pdf.LeftMargin + textBlockWidth

	pdf.Bookmark(title, 0, -1)

	pdf.Ln(1.5)
	headingHeight, _ := pdf.TextCell(textBlockWidth, 6.5, title, 0, "LT", 18, "B", 1.5, "")
	pdf.Ln((headingHeight - 1.5) * 1.3)
	_, descFontSize := pdf.TextCell(textBlockWidth, 5, description1, -1, "LT", 10, "", 1.5, description)
	pdf.TextCell(pdf.GetStringWidth(description2), 5, description2, 0, "LT", descFontSize, "I", 0, description2)

	pdf.Line(qrBlockX, pdf.TopMargin, qrBlockX, pdf.TopMargin+qrBlockDimensions)
	qrKey := barcode.RegisterQR(pdf, path, qr.L, qr.Unicode)
	barcode.Barcode(
		pdf, qrKey,
		pdf.PageWidth-pdf.RightMargin-qrBlockDimensions+1,
		pdf.TopMargin+1,
		13, 13, false,
	)
	pdf.Line(pdf.LeftMargin, pdf.TopMargin+qrBlockDimensions, pdf.PageWidth-pdf.RightMargin, pdf.TopMargin+qrBlockDimensions)
	pdf.Ln(5.7)
}

func (pdf PDF) addTableHeader(rowHeight float64) {
	pdf.SetFont(pdf.FontFamily, "B", 7)
	pdf.SetCellMargin(1.5)
	pdf.CellFormat(23, rowHeight, "Beleg", "TB", 0, "L", false, 0, "")
	pdf.SetCellMargin(0)
	pdf.CellFormat(14, rowHeight, "Datum", "TB", 0, "L", false, 0, "")
	pdf.CellFormat(103.49, rowHeight, "Beschreibung", "TB", 0, "L", false, 0, "")
	pdf.CellFormat(14, rowHeight, "Soll", "TB", 0, "R", false, 0, "")
	pdf.CellFormat(14, rowHeight, "Haben", "TB", 0, "R", false, 0, "")
	pdf.CellFormat(20, rowHeight, "Betrag", "TB", 1, "R", false, 0, "")
}

func (pdf PDF) addTableRows(transactions Transactions, rowHeight float64) {
	pdf.SetFont(pdf.FontFamily, "", 7)

	// Merged cell for "Ident"
	pdf.SetCellMargin(1.5)
	pdf.CellFormat(23, rowHeight, "<Ident>", "T", 0, "L", false, 0, "")
	pdf.SetCellMargin(0)

	// First row of the group
	first := true
	for _, tx := range transactions {
		if tx.Ident == "e-inv-4" {
			fmt.Println(strconv.Quote(tx.FmtDescription()))
		}

		if !first {
			// Empty Ident cell for subsequent rows
			pdf.CellFormat(23, rowHeight, "", "", 0, "L", false, 0, "")
		}
		// Add transaction details
		pdf.TableCell(14, rowHeight, tx.FmtDate(), "B", 0, "L")
		pdf.TableCell(103.49, rowHeight, tx.FmtDescription(), "B", 0, "L")
		pdf.TableCell(14, rowHeight, tx.AccountDebit, "B", 0, "R")
		pdf.TableCell(14, rowHeight, tx.AccountCredit, "B", 0, "R")
		pdf.TableCell(20, rowHeight, tx.FmtAmount(), "B", 1, "R")
		first = false
		pdf.SetDashPattern([]float64{}, 0)
	}

	// Draw dotted separator
	pdf.SetDrawColor(150, 150, 150)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.SetDrawColor(0, 0, 0)
}

func (pdf PDF) Lol(document Document) {
	pdf.SetFont(pdf.FontFamily, "", 7)
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

func (pdf PDF) TextCell(
	w, h float64,
	txtStr string,
	ln int,
	alignStr string,
	size float64,
	style string,
	margin float64,
	calcTxtStr string,
) (float64, float64) {
	if style != "" && style != "I" && style != "B" {
		panic(fmt.Sprintf("text style '%s' is not supported", style))
	}

	if calcTxtStr == "" {
		calcTxtStr = txtStr
	}
	maxTextWidth := w - 2*margin
	pdf.SetFont(pdf.FontFamily, style, size)
	for pdf.GetStringWidth(calcTxtStr) > maxTextWidth {
		size -= .5
		pdf.SetFont(pdf.FontFamily, style, size)
	}
	// Set width to actual needed width
	w = pdf.GetStringWidth(txtStr) + 1.5

	if ln == -1 {
		pdf.Ln(margin)
	}
	pdf.SetCellMargin(margin)

	drawR, drawG, drawB := pdf.GetDrawColor()
	borderStr := ""
	if pdf.debugMode {
		borderStr = "1"
		pdf.setDebugDrawColor()
	}

	pdf.CellFormat(w, h, txtStr, borderStr, ln, alignStr, false, 0, "")

	if pdf.debugMode {
		pdf.SetDrawColor(drawR, drawG, drawB)
	}

	_, fontHeight := pdf.GetFontSize()
	return fontHeight, size
}

func (pdf PDF) TableCell(w, h float64, txtStr string, borderStr string, ln int, alignStr string) {
	drawR, drawG, drawB := pdf.GetDrawColor()
	if pdf.debugMode {
		borderStr = "1"
		pdf.setDebugDrawColor()
	}

	for pdf.GetStringWidth(txtStr) > w-1.5 {
		txtStr = strings.TrimSuffix(txtStr, "…")
		txtStr = txtStr[:len(txtStr)-1]
		txtStr = txtStr + "…"
	}

	pdf.CellFormat(w, h, txtStr, borderStr, ln, alignStr, false, 0, "")

	if pdf.debugMode {
		pdf.SetDrawColor(drawR, drawG, drawB)
	}
}

func (pdf PDF) setDebugDrawColor() {
	pdf.SetDrawColor(0xDA, 0x6A, 0x35)
}
