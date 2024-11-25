package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/go-pdf/fpdf"
	cbarcode "github.com/go-pdf/fpdf/contrib/barcode"
	"github.com/go-pdf/fpdf/contrib/gofpdi"
)

//go:embed static/literata-regular.ttf
var literataRegular []byte

//go:embed static/literata-italic.ttf
var literataItalic []byte

//go:embed static/literata-medium.ttf
var literataMedium []byte

//go:embed static/sad-document.png
var sadDocument []byte

type DebugColor int

const (
	ColorMagenta DebugColor = iota
	ColorTeal
	ColorGreen
	ColorVermilion
	ColorAmberLight
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
	case ColorAmberLight:
		return 0xf9, 0xdd, 0x9d
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
	pdf.AddUTF8FontFromBytes(fontName, "B", literataMedium)

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
		if doc.Path != "../internal-expenses/2023/hetzner_2023-10-01_R0020566025.pdf" {
			continue
		}
		embedPDFPageCount := 1
		for page := 1; page <= embedPDFPageCount; page++ {
			embedPDFPageCount = pdf.addDocument(*dossier, doc, page, i+1)
		}
		if i == 10 {
			// FOR DEBUG
			// return
		}
	}
}

func (pdf PDF) addDocument(dossier Dossier, doc Document, embedPageNr, reportPageCount int) (pageCount int) {
	// If not -1 there is another page from this receipt.
	pdf.AddPage()
	pdf.Rect(pdf.LeftMargin, pdf.TopMargin, pdf.AreaWidth, pdf.AreaHeight, "D")
	pdf.addHeader(doc, embedPageNr)

	if embedPageNr == 1 {
		pdf.addTableHeader(5)
		pdf.addTableRows(doc.Transactions, 5, dossier.BaseCurrency)
	}
	pageCount = pdf.embedDocument(dossier, doc, embedPageNr, 10)
	pdf.addFooter(dossier, doc, 10, embedPageNr, pageCount, reportPageCount)
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
	headingHeight, _ := pdf.TextCell(textBlockWidth, 6.5, title, 0, "LT", 18, "B", 1.5, "", true)
	pdf.Ln((headingHeight - 1.5) * 1.3)
	_, descFontSize := pdf.TextCell(textBlockWidth, 5, description1, -1, "LT", 10, "", 1.5, description, true)
	pdf.TextCell(pdf.GetStringWidth(description2), 5, description2, 0, "LT", descFontSize, "B", 0, description2, true)

	pdf.Line(qrBlockX, pdf.TopMargin, qrBlockX, pdf.TopMargin+qrBlockDimensions)
	qrCode, err := qr.Encode(doc.Path, qr.L, qr.Unicode)
	if err != nil {
		panic(err)
	}
	qrCode, err = barcode.Scale(qrCode, 256, 256)
	if err != nil {
		panic(err)
	}
	qrKey := cbarcode.Register(qrCode)
	cbarcode.Barcode(
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

func (pdf PDF) addTableRows(transactions Transactions, rowHeight float64, baseCurrency string) {
	pdf.SetFont(pdf.FontFamily, "", 7)

	// First row of the group
	first := true
	previousIdent := ""
	for _, tx := range transactions {

		if previousIdent != tx.Ident {
			if !first {
				pdf.HLine(0, false, ColorMagenta)
			}

			pdf.SetCellMargin(1.5)
			pdf.CellFormat(23, rowHeight, tx.Ident, "", 0, "L", tx.Ident == "", 0, "")
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

		if tx.ExchangeCurrency != baseCurrency {
			pdf.ForeignAmountTableCell(20, rowHeight, tx, baseCurrency)
		} else {
			amount := fmt.Sprint(tx.Amount, " ", baseCurrency)
			pdf.TableCell(20, rowHeight, amount, "", 0, "R")
		}

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

func (pdf PDF) addFooter(dossier Dossier, doc Document, footerHeight float64, embedPageNr, embedTotalPages, reportPageCount int) {
	startY := pdf.AreaHeight + pdf.TopMargin - footerHeight
	if pdf.debugLines {
		pdf.SetDrawColor(255, 0, 255)
	}
	pdf.Line(pdf.LeftMargin, startY, pdf.PageWidth-pdf.RightMargin, startY)
	if pdf.debugLines {
		pdf.SetDrawColor(0, 0, 0)
	}

	lastAddressLine := fmt.Sprintf("%s %s", dossier.ZIPCode, dossier.Place)
	if embedTotalPages == 0 {
		embedTotalPages = 1
	}
	countInfo := fmt.Sprintf("%d/%d – Page %d", embedPageNr, embedTotalPages, reportPageCount)
	file := fmt.Sprint("File: ", filepath.Base(dossier.AccountingFilePath))
	lastSaved := fmt.Sprint("Accounting data as of: ", dossier.FmtLastSaved())
	createdAt := fmt.Sprint("Report was created on: ", time.Now().Format(DATE_TIME_FORMAT))

	pdf.SetY(startY + .5)
	lineHeight := (footerHeight - 1) / 3
	cellWidth := pdf.AreaWidth / 3
	size := 6.8
	pdf.TextCell(cellWidth, lineHeight, dossier.CompanyName, 0, "LM", size, "", .5, "", false)
	pdf.TextCell(cellWidth, lineHeight, countInfo, 0, "CM", size, "B", .5, "", false)
	pdf.TextCell(cellWidth, lineHeight, file, 1, "RM", size, "", .5, "", false)
	pdf.TextCell(cellWidth, lineHeight, dossier.Street, 0, "LM", size, "", .5, "", false)
	pdf.TextCell(cellWidth, lineHeight, "", 0, "CM", size, "", .5, "", false)
	pdf.TextCell(cellWidth, lineHeight, lastSaved, 1, "RM", size, "", .5, "", false)
	pdf.TextCell(cellWidth, lineHeight, lastAddressLine, 0, "LM", size, "", .5, "", false)
	pdf.TextCell(cellWidth, lineHeight, dossier.FmtPeriod(), 0, "CM", size, "", .5, "", false)
	pdf.TextCell(cellWidth, lineHeight, createdAt, 0, "RM", size, "", .5, "", false)
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
	shrinkToTxt bool,
) (float64, float64) {
	if style != "" && style != "I" && style != "B" {
		panic(fmt.Sprintf("text style '%s' is not supported", style))
	}

	if calcTxtStr == "" {
		calcTxtStr = txtStr
	}
	size = pdf.fitTextToWidth(calcTxtStr, w, margin, size, style)
	if shrinkToTxt {
		w = pdf.GetStringWidth(txtStr) + 1.5
	}

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
	if pdf.debugCells {
		borderStr = "1"
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

func (pdf PDF) ForeignAmountTableCell(
	w, h float64,
	transaction Transaction,
	baseCurrency string,
) {
	drawR, drawG, drawB := pdf.setDebugDrawColor(pdf.debugCells, ColorVermilion)
	borderStr := ""
	if pdf.debugCells {
		borderStr = "1"
	}
	fontSize, _ := pdf.GetFontSize()

	pdf.SetFont(pdf.FontFamily, "", 7)
	_, height7pt := pdf.GetFontSize()
	pdf.SetFont(pdf.FontFamily, "", 4)
	_, height5pt := pdf.GetFontSize()
	margin := (h - height7pt - height5pt) / 2

	pdf.SetFont(pdf.FontFamily, "", 7)
	baseAmount := fmt.Sprint(transaction.Amount, " ", baseCurrency)
	pdf.CellFormat(w, height7pt+margin, baseAmount, borderStr, 2, "RB", false, 0, "")

	exchangeInfo := fmt.Sprintf(
		"%s %s – %s",
		transaction.AmountCurrency,
		transaction.ExchangeCurrency,
		transaction.ExchangeRate[:6],
	)
	pdf.SetFont(pdf.FontFamily, "", 4)
	pdf.CellFormat(w, height5pt+margin, exchangeInfo, borderStr, 2, "RT", false, 0, "")

	if pdf.debugCells {
		pdf.SetDrawColor(drawR, drawG, drawB)
	}
	pdf.SetFontSize(fontSize)
	pdf.CellFormat(0, 0, "", "", 1, "", false, 0, "")
}

func (pdf PDF) MultilineTextCell(
	w float64,
	lineSpread float64,
	txtStr string,
	alignStr string,
	size float64,
	style string,
	margin float64,
) {
	if style != "" && style != "I" && style != "B" {
		panic(fmt.Sprintf("text style '%s' is not supported", style))
	}

	drawR, drawG, drawB := pdf.setDebugDrawColor(pdf.debugCells, ColorVermilion)
	borderStr := ""
	if pdf.debugCells {
		borderStr = "1"
	}

	pdf.SetFont(pdf.FontFamily, style, size)
	_, fontHeight := pdf.GetFontSize()
	pdf.Ln(margin)
	pdf.SetCellMargin(margin)

	pdf.MultiCell(w, fontHeight*lineSpread, txtStr, borderStr, alignStr, false)

	if pdf.debugCells {
		pdf.SetDrawColor(drawR, drawG, drawB)
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
