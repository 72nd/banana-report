package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/go-pdf/fpdf"
	fc "github.com/go-pdf/fpdf/contrib/barcode"
	"github.com/go-pdf/fpdf/contrib/gofpdi"
)

//go:embed literata_7pt-regular.ttf
var literataRegular []byte

//go:embed literata_7pt-italic.ttf
var literataItalic []byte

//go:embed literata_7pt-semi-bold.ttf
var literataSemiBold []byte

type DebugColor int

const (
	ColorMagenta DebugColor = iota
	ColorTeal
	ColorGreen
)

type PDF struct {
	*fpdf.Fpdf
	debugCells   bool
	debugLines   bool
	PageWidth    float64
	PageHeight   float64
	AreaWidth    float64
	AreaHeight   float64
	FontFamily   string
	LeftMargin   float64
	TopMargin    float64
	RightMargin  float64
	BottomMargin float64
}

func NewPDF(debugCells bool, debugLines bool) PDF {
	fontName := "Literata"
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(false, 10)
	pdf.AddUTF8FontFromBytes(fontName, "", literataRegular)
	pdf.AddUTF8FontFromBytes(fontName, "I", literataItalic)
	pdf.AddUTF8FontFromBytes(fontName, "B", literataSemiBold)

	pageWidth, pageHeight := pdf.GetPageSize()
	lm, tm, rm, bm := pdf.GetMargins()
	return PDF{
		Fpdf:         pdf,
		debugCells:   debugCells,
		debugLines:   debugLines,
		PageWidth:    pageWidth,
		PageHeight:   pageHeight,
		AreaWidth:    pageWidth - lm - rm,
		AreaHeight:   pageHeight - tm - bm,
		FontFamily:   fontName,
		LeftMargin:   lm,
		TopMargin:    tm,
		RightMargin:  rm,
		BottomMargin: bm,
	}
}

func (pdf PDF) Build(dossier *Dossier) {
	for i, doc := range dossier.JournalEntries {
		pdf.addDocument(*dossier, doc)
		if i == 4 {
			return
		}
	}
}

func (pdf PDF) addDocument(dossier Dossier, doc Document) {
	pdf.AddPage()
	pdf.Rect(pdf.LeftMargin, pdf.TopMargin, pdf.AreaWidth, pdf.AreaHeight, "D")
	pdf.addHeader(doc)

	pdf.addTableHeader(4.5)
	tableBottomY := pdf.addTableRows(doc.Transactions, 4.5)
	pdf.embedDocument(dossier, doc, tableBottomY, 10)
	pdf.addFooter(doc, 0)
}

func (pdf PDF) addHeader(doc Document) {
	title := strings.TrimSuffix(filepath.Base(doc.Path), filepath.Ext(doc.Path))
	description1 := fmt.Sprintf("%s — ", doc.Path)
	description2 := doc.IdentStringList()
	description := fmt.Sprintf("%s%s", description1, description2)
	qrBlockDimensions := 15.
	textBlockWidth := pdf.AreaWidth - qrBlockDimensions
	qrBlockX := pdf.LeftMargin + textBlockWidth

	pdf.Bookmark(title, 0, -1)

	pdf.Ln(1.5)
	headingHeight, _ := pdf.TextCell(textBlockWidth, 6.5, title, 0, "LT", 18, "B", 1.5, "")
	pdf.Ln((headingHeight - 1.5) * 1.3)
	_, descFontSize := pdf.TextCell(textBlockWidth, 5, description1, -1, "LT", 10, "", 1.5, description)
	pdf.TextCell(pdf.GetStringWidth(description2), 5, description2, 0, "LT", descFontSize, "I", 0, description2)

	pdf.Line(qrBlockX, pdf.TopMargin, qrBlockX, pdf.TopMargin+qrBlockDimensions)
	qrCode, err := qr.Encode(doc.Path, qr.L, qr.Unicode)
	if err != nil {
		panic(err)
	}
	qrCode, err = barcode.Scale(qrCode, 256, 256)
	if err != nil {
		panic(err)
	}
	qrKey := fc.Register(qrCode)
	fc.Barcode(
		pdf, qrKey,
		pdf.PageWidth-pdf.RightMargin-qrBlockDimensions+.5,
		pdf.TopMargin+.5,
		14, 14, false,
	)
	pdf.Line(pdf.LeftMargin, pdf.TopMargin+qrBlockDimensions, pdf.PageWidth-pdf.RightMargin, pdf.TopMargin+qrBlockDimensions)
	pdf.Ln(5.7)
}

func (pdf PDF) addTableHeader(rowHeight float64) {
	pdf.SetFont(pdf.FontFamily, "B", 7)
	pdf.SetCellMargin(1.5)
	pdf.CellFormat(23, rowHeight, "Beleg", "", 0, "L", false, 0, "")
	pdf.SetCellMargin(0)
	pdf.CellFormat(14, rowHeight, "Datum", "", 0, "L", false, 0, "")
	pdf.CellFormat(103.49, rowHeight, "Beschreibung", "", 0, "L", false, 0, "")
	pdf.CellFormat(14, rowHeight, "Soll", "", 0, "R", false, 0, "")
	pdf.CellFormat(14, rowHeight, "Haben", "", 0, "R", false, 0, "")
	pdf.CellFormat(20, rowHeight, "Betrag", "", 1, "R", false, 0, "")
	pdf.HLine(0, false, ColorTeal)
}

func (pdf PDF) addTableRows(transactions Transactions, rowHeight float64) float64 {
	pdf.SetFont(pdf.FontFamily, "", 7)

	// First row of the group
	first := true
	previousIdent := ""
	for _, tx := range transactions {

		if previousIdent != tx.Ident {
			// Merged cell for "Ident"
			if !first {
				pdf.HLine(0, false, ColorMagenta)
			}

			pdf.SetCellMargin(1.5)
			pdf.CellFormat(23, rowHeight, tx.Ident, "", 0, "L", false, 0, "")
			pdf.SetCellMargin(0)
		} else {
			// Empty Ident cell for subsequent rows
			pdf.HLine(23, true, ColorGreen)
			pdf.CellFormat(23, rowHeight, "", "", 0, "L", false, 0, "")
		}

		// Add transaction details
		pdf.TableCell(14, rowHeight, tx.FmtDate(), "", 0, "L")
		pdf.TableCell(103.49, rowHeight, tx.FmtDescription(), "", 0, "L")
		pdf.TableCell(14, rowHeight, tx.AccountDebit, "", 0, "R")
		pdf.TableCell(14, rowHeight, tx.AccountCredit, "", 0, "R")
		pdf.TableCell(20, rowHeight, tx.FmtAmount(), "", 1, "R")
		pdf.SetDashPattern([]float64{}, 0)
		first = false
		previousIdent = tx.Ident
	}
	pdf.HLine(0, false, ColorMagenta)
	return pdf.GetY()
}

func (pdf PDF) embedDocument(dossier Dossier, doc Document, tableBottomY, footerHeight float64) {
	path, err := dossier.ResolveRelativePath(doc.Path)
	if err != nil {
		// TODO: Print error to the document.
		panic(err)
	}
	if _, err = os.Stat(path); err != nil {
		panic(err)
	}
	tpl := gofpdi.ImportPage(pdf, path, 1, "/MediaBox")
	width, height := fitImage(
		210,
		297,
		pdf.AreaWidth,
		pdf.AreaHeight-tableBottomY-pdf.BottomMargin-footerHeight,
	)
	gofpdi.UseImportedTemplate(pdf, tpl, pdf.LeftMargin+1.5, tableBottomY+1.5, width, height)
}

func (pdf PDF) addFooter(doc Document, footerHeight float64) {
	lineY := pdf.AreaHeight - footerHeight
	if pdf.debugLines {
		pdf.SetDrawColor(255, 0, 255)
	}
	pdf.Line(pdf.LeftMargin, lineY, pdf.PageWidth-pdf.RightMargin, lineY)
	if pdf.debugLines {
		pdf.SetDrawColor(0, 0, 0)
	}
}

func (pdf PDF) HLine(x1 float64, dotted bool, debugColor DebugColor) {
	if pdf.debugLines && debugColor == ColorTeal {
		// Teal
		pdf.SetDrawColor(0x43, 0x95, 0xb7)
	} else if pdf.debugLines && debugColor == ColorGreen {
		// Green
		pdf.SetDrawColor(0x3e, 0x8c, 0x5f)
	} else if pdf.debugLines {
		// Magenta
		pdf.SetDrawColor(255, 0, 255)
	}

	if dotted {
		pdf.SetDashPattern([]float64{.6, .6}, 0)
	}

	pdf.Line(pdf.LeftMargin+x1, pdf.GetY(), pdf.PageWidth-pdf.RightMargin, pdf.GetY())

	if dotted {
		pdf.SetDashPattern([]float64{}, 0)
	}
	if pdf.debugLines {
		pdf.SetDrawColor(0, 0, 0)
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
	if pdf.debugCells {
		borderStr = "1"
		pdf.setDebugDrawColor()
	}

	pdf.CellFormat(w, h, txtStr, borderStr, ln, alignStr, false, 0, "")

	if pdf.debugCells {
		pdf.SetDrawColor(drawR, drawG, drawB)
	}

	_, fontHeight := pdf.GetFontSize()
	return fontHeight, size
}

func (pdf PDF) TableCell(w, h float64, txtStr string, borderStr string, ln int, alignStr string) {
	drawR, drawG, drawB := pdf.GetDrawColor()
	if pdf.debugCells {
		borderStr = "1"
		pdf.setDebugDrawColor()
	}

	for pdf.GetStringWidth(txtStr) > w-1.5 {
		txtStr = strings.TrimSuffix(txtStr, "…")
		txtStr = txtStr[:len(txtStr)-1]
		txtStr = txtStr + "…"
	}

	pdf.CellFormat(w, h, txtStr, borderStr, ln, alignStr, false, 0, "")

	if pdf.debugCells {
		pdf.SetDrawColor(drawR, drawG, drawB)
	}
}

func (pdf PDF) setDebugDrawColor() {
	pdf.SetDrawColor(0xDA, 0x6A, 0x35)
}

func fitImage(origWidth, origHeight, maxWidth, maxHeight float64) (float64, float64) {
	if origWidth <= 0 || origHeight <= 0 || maxWidth <= 0 || maxHeight <= 0 {
		return 0, 0
	}
	aspectRatio := origWidth / origHeight

	var newWidth, newHeight float64
	if maxWidth/maxHeight > aspectRatio {
		newHeight = maxHeight
		newWidth = aspectRatio * maxHeight
	} else {
		newWidth = maxWidth
		newHeight = maxWidth / aspectRatio
	}
	return newWidth, newHeight
}
