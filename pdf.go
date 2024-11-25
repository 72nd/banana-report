package main

import (
	"bytes"
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

//go:embed static/literata_7pt-regular.ttf
var literataRegular []byte

//go:embed static/literata_7pt-italic.ttf
var literataItalic []byte

//go:embed static/literata_7pt-semi-bold.ttf
var literataSemiBold []byte

//go:embed static/sad-document.png
var sadDocument []byte

type DebugColor int

const (
	ColorMagenta DebugColor = iota
	ColorTeal
	ColorGreen
	ColorVermilion
)

func (c DebugColor) GetValues() (r, g, b int) {
	// Defaults to magenta.
	switch c {
	case ColorTeal:
		return 0x43, 0x95, 0xb7
	case ColorGreen:
		return 0x3e, 0x8c, 0x5f
	case ColorVermilion:
		return 0xDA, 0x6A, 0x35
	default:
		// Magenta
		return 0xB4, 0x25, 0x7A
	}
}

type PDF struct {
	*fpdf.Fpdf
	debugCells         bool
	debugLines         bool
	sadDocumentOptions fpdf.ImageOptions
	PageWidth          float64
	PageHeight         float64
	AreaWidth          float64
	AreaHeight         float64
	FontFamily         string
	LeftMargin         float64
	TopMargin          float64
	RightMargin        float64
	BottomMargin       float64
}

func NewPDF(debugCells bool, debugLines bool) PDF {
	fontName := "Literata"
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(false, 10)
	pdf.AddUTF8FontFromBytes(fontName, "", literataRegular)
	pdf.AddUTF8FontFromBytes(fontName, "I", literataItalic)
	pdf.AddUTF8FontFromBytes(fontName, "B", literataSemiBold)

	imgRd := bytes.NewReader(sadDocument)
	sadDocumentOpt := fpdf.ImageOptions{
		ImageType: "PNG",
	}
	pdf.RegisterImageOptionsReader("sad-document", sadDocumentOpt, imgRd)

	pageWidth, pageHeight := pdf.GetPageSize()
	lm, tm, rm, bm := pdf.GetMargins()
	return PDF{
		Fpdf:               pdf,
		debugCells:         debugCells,
		debugLines:         debugLines,
		sadDocumentOptions: sadDocumentOpt,
		PageWidth:          pageWidth,
		PageHeight:         pageHeight,
		AreaWidth:          pageWidth - lm - rm,
		AreaHeight:         pageHeight - tm - bm,
		FontFamily:         fontName,
		LeftMargin:         lm,
		TopMargin:          tm,
		RightMargin:        rm,
		BottomMargin:       bm,
	}
}

func (pdf PDF) Build(dossier *Dossier) {
	for i, doc := range dossier.JournalEntries {
		embedPDFPageCount := 1
		for page := 1; page <= embedPDFPageCount; page++ {
			embedPDFPageCount = pdf.addDocument(*dossier, doc, page)
		}
		if i == 10 {
			// return
		}
	}
}

func (pdf PDF) addDocument(dossier Dossier, doc Document, embedPageNr int) (pageCount int) {
	// If not -1 there is another page from this receipt.
	pdf.AddPage()
	pdf.Rect(pdf.LeftMargin, pdf.TopMargin, pdf.AreaWidth, pdf.AreaHeight, "D")
	pdf.addHeader(doc, embedPageNr)

	if embedPageNr == 1 {
		pdf.addTableHeader(4.5)
		pdf.addTableRows(doc.Transactions, 4.5)
	}
	pageCount = pdf.embedDocument(dossier, doc, embedPageNr, 10)
	pdf.addFooter(doc, 10)
	return pageCount
}

func (pdf PDF) addHeader(doc Document, embedPageNr int) {
	title := strings.TrimSuffix(filepath.Base(doc.Path), filepath.Ext(doc.Path))
	if embedPageNr != 1 {
		title = fmt.Sprintf("%s (cont.)", title)
	}
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
	endOfHeaderY := pdf.TopMargin + qrBlockDimensions
	pdf.Line(pdf.LeftMargin, endOfHeaderY, pdf.PageWidth-pdf.RightMargin, pdf.TopMargin+qrBlockDimensions)
	pdf.SetY(endOfHeaderY)
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

func (pdf PDF) addTableRows(transactions Transactions, rowHeight float64) {
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
}

type EmbedError struct {
	Operation string
	Error     error
}

func (e EmbedError) String() string {
	return fmt.Sprintf("Error occurred during %s: %s.", e.Operation, e.Error)
}

func (pdf PDF) embedDocument(dossier Dossier, doc Document, page int, footerHeight float64) (pageCount int) {
	errors := []EmbedError{}
	path, err := dossier.ResolveRelativePath(doc.Path)
	if err != nil {
		errors = append(errors, EmbedError{
			Operation: "resolve absolute path",
			Error:     err,
		})
	}
	if _, err = os.Stat(path); err != nil {
		errors = append(errors, EmbedError{
			Operation: "check file existence",
			Error:     err,
		})
	}
	var embedErr error
	pageCount = pdf.embedPDF(path, 1, pdf.GetY(), footerHeight, &embedErr)
	if embedErr != nil {
		errors = append(errors, EmbedError{
			Operation: "embedding file",
			Error:     embedErr,
		})
	}
	if len(errors) != 0 {
		pdf.addEmbedPDFErrors(errors)
		for _, err := range errors {
			fmt.Printf("%s: error during %s\n", path, err.Operation)
		}
	}
	return pageCount
}

func (pdf PDF) embedPDF(path string, page int, tableBottomY, footerHeight float64, err *error) (totalPages int) {
	defer func() {
		if r := recover(); r != nil {
			*err = fmt.Errorf("'%s', try to reexport the file in order to fix it", r)
		}
	}()
	tpl := gofpdi.ImportPage(pdf, path, 1, "/MediaBox")
	width, height := fitImage(
		210,
		297,
		pdf.AreaWidth-2,
		pdf.AreaHeight-tableBottomY+footerHeight-footerHeight-2,
	)
	x := (pdf.AreaWidth - 2 - width) / 2
	gofpdi.UseImportedTemplate(pdf, tpl, pdf.LeftMargin+x+1, tableBottomY+1, width, height)

	return len(gofpdi.GetPageSizes())
}

func (pdf PDF) addEmbedPDFErrors(errors []EmbedError) {
	pdf.ImageOptions(
		"sad-document",
		(pdf.AreaWidth-20)/2,
		pdf.GetY()+5,
		40, 40, false, pdf.sadDocumentOptions, 0, "",
	)

	errStr := []string{}
	for _, err := range errors {
		errStr = append(errStr, fmt.Sprintf("- %s", err))
	}
	txt := fmt.Sprintf("One or more error(s) occurred during embedding the file:\n\n%s", strings.Join(errStr, "\n"))

	pdf.SetY(pdf.GetY() + 40 + 10)
	pdf.SetCellMargin(1.5)
	pdf.SetFont(pdf.FontFamily, "", 11)

	drawR, drawG, drawB := pdf.setDebugDrawColor(pdf.debugCells, ColorVermilion)
	borderStr := ""
	if pdf.debugCells {
		borderStr = "1"
	}
	pdf.MultiCell(pdf.AreaWidth, 5, txt, borderStr, "LT", false)
	if pdf.debugCells {
		pdf.SetDrawColor(drawR, drawG, drawB)
	}
}

func (pdf PDF) addFooter(doc Document, footerHeight float64) {
	lineY := pdf.AreaHeight + pdf.TopMargin - footerHeight
	if pdf.debugLines {
		pdf.SetDrawColor(255, 0, 255)
	}
	pdf.Line(pdf.LeftMargin, lineY, pdf.PageWidth-pdf.RightMargin, lineY)
	if pdf.debugLines {
		pdf.SetDrawColor(0, 0, 0)
	}
}

func (pdf PDF) HLine(x1 float64, dotted bool, debugColor DebugColor) {
	drawR, drawG, drawB := pdf.setDebugDrawColor(pdf.debugLines, debugColor)

	if dotted {
		pdf.SetDashPattern([]float64{.6, .6}, 0)
	}

	pdf.Line(pdf.LeftMargin+x1, pdf.GetY(), pdf.PageWidth-pdf.RightMargin, pdf.GetY())

	if dotted {
		pdf.SetDashPattern([]float64{}, 0)
	}
	if pdf.debugLines {
		pdf.SetDrawColor(drawR, drawG, drawB)
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
	size = pdf.fitTextToWidth(calcTxtStr, w, margin, size, style)
	w = pdf.GetStringWidth(txtStr) + 1.5

	if ln == -1 {
		pdf.Ln(margin)
	}
	pdf.SetCellMargin(margin)

	drawR, drawG, drawB := pdf.setDebugDrawColor(pdf.debugCells, ColorVermilion)
	borderStr := ""
	if pdf.debugCells {
		borderStr = "1"
	}

	pdf.CellFormat(w, h, txtStr, borderStr, ln, alignStr, false, 0, "")

	if pdf.debugCells {
		pdf.SetDrawColor(drawR, drawG, drawB)
	}

	_, fontHeight := pdf.GetFontSize()
	return fontHeight, size
}

func (pdf PDF) TableCell(w, h float64, txtStr string, borderStr string, ln int, alignStr string) {
	drawR, drawG, drawB := pdf.setDebugDrawColor(pdf.debugCells, ColorVermilion)

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

func (pdf PDF) MultilineTextCell(
	w float64,
	lh float64,
	txtStr string,
	alignStr string,
	size float64,
	style string,
	margin float64,
) {
	if style != "" && style != "I" && style != "B" {
		panic(fmt.Sprintf("text style '%s' is not supported", style))
	}

}

func (pdf PDF) setDebugDrawColor(isDebugEnabled bool, color DebugColor) (r, g, b int) {
	r, g, b = pdf.GetDrawColor()
	if !isDebugEnabled {
		return
	}
	pdf.SetDrawColor(color.GetValues())
	return r, g, b
}

func (pdf PDF) fitTextToWidth(txt string, width, margin, maxSize float64, style string) float64 {
	// Returns font size.
	rsl := maxSize
	maxTextWidth := width - 2*margin
	pdf.SetFont(pdf.FontFamily, style, maxSize)
	for pdf.GetStringWidth(txt) > maxTextWidth {
		rsl -= .5
		pdf.SetFont(pdf.FontFamily, style, rsl)
	}
	return rsl
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
