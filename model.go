package main

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const DATE_FORMAT = "02.01.2006"
const TIME_FORMAT = "15:04:05"
const DATE_TIME_FORMAT = "02.01.06 15:04:05"

var CURRENCY_SYMBOLS = map[string]string{
	"EUR": "€",
	"USD": "$",
}

// Get the currency symbol for the given currency code.
func getCurrencySymbol(currency string) string {
	symbol, ok := CURRENCY_SYMBOLS[currency]
	if !ok {
		return currency
	}
	return symbol
}

// Everything with the same linked document. Filepath is map key.
type Dossier struct {
	JournalEntries     Documents
	AccountingFilePath string
	BaseCurrency       string
	CompanyName        string
	Street             string
	ZIPCode            string
	Place              string
	DateLastSaved      time.Time
	TimeLastSaved      time.Time
	OpeningDate        time.Time
	ClosureDate        time.Time
}

func DossierFromXML(path string) (*Dossier, error) {
	ac, err := AC2FromFile(path)
	if err != nil {
		return nil, err
	}

	journalTable, err := ac.TableById("Journal")
	if err != nil {
		return nil, err
	}
	journal := JournalFromTable(*journalTable)
	entries := EntriesFromJournal(journal)

	fileInfoTable, err := ac.TableById("FileInfo")
	if err != nil {
		return nil, err
	}

	return &Dossier{
		JournalEntries:     entries,
		AccountingFilePath: fileInfoTable.GuardedValueById("FileName"),
		BaseCurrency:       getCurrencySymbol(fileInfoTable.GuardedValueById("BasicCurrency")),
		CompanyName:        fileInfoTable.GuardedValueById("Company"),
		Street:             fileInfoTable.GuardedValueById("Address1"),
		ZIPCode:            fileInfoTable.GuardedValueById("Zip"),
		Place:              fileInfoTable.GuardedValueById("City"),
		DateLastSaved:      fileInfoTable.GuardedDateById("DateLastSaved"),
		TimeLastSaved:      fileInfoTable.GuardedTimeById("TimeLastSaved"),
		OpeningDate:        fileInfoTable.GuardedDateById("OpeningDate"),
		ClosureDate:        fileInfoTable.GuardedDateById("ClosureDate"),
	}, nil
}

func (d Dossier) ResolveRelativePath(path string) (string, error) {
	fullPath := filepath.Join(filepath.Dir(d.AccountingFilePath), path)
	return filepath.Abs(fullPath)
}

func (d Dossier) FmtLastSaved() string {
	dt := d.DateLastSaved.Format(DATE_FORMAT)
	if d.DateLastSaved.Equal(time.Time{}) {
		dt = UNKNOWN_STR
	}
	tm := d.TimeLastSaved.Format(TIME_FORMAT)
	if d.TimeLastSaved.Equal(time.Time{}) {
		tm = UNKNOWN_STR
	}
	return fmt.Sprint(dt, " ", tm)
}

func (d Dossier) FmtPeriod() string {
	from := d.OpeningDate.Format(DATE_FORMAT)
	if d.OpeningDate.Equal(time.Time{}) {
		from = UNKNOWN_STR
	}
	to := d.ClosureDate.Format(DATE_FORMAT)
	if d.ClosureDate.Equal(time.Time{}) {
		to = UNKNOWN_STR
	}
	return fmt.Sprint(from, " – ", to)
}

type Documents []Document

func EntriesFromJournal(journal Transactions) Documents {
	tmp := map[string]Document{}
	for _, transaction := range journal {
		path := strings.TrimSpace(transaction.Path)
		if _, exists := tmp[path]; !exists {
			tmp[path] = Document{
				Path:         path,
				Transactions: []Transaction{},
			}
		}
		// fmt.Println(tmp[path][transaction.Ident].transactions)
		doc := tmp[path]
		doc.Transactions = append(doc.Transactions, transaction)
		tmp[path] = doc
	}
	rsl := Documents{}
	for _, doc := range tmp {
		sort.Sort(doc.Transactions)
		rsl = append(rsl, doc)
	}
	sort.Sort(rsl)
	return rsl
}

func (d Documents) Len() int {
	return len(d)
}

func (d Documents) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d Documents) Less(i, j int) bool {
	return path.Base(d[i].Path) < path.Base(d[j].Path)
}

// All docs of one doc-ident. Doc-ident is map key.
type Document struct {
	Path         string
	Transactions Transactions
}

func (d Document) IdentStringList() string {
	rsl := []string{}
	for _, transaction := range d.Transactions {
		add := true
		for _, item := range rsl {
			if item == transaction.Ident {
				add = false
				break
			}
		}
		if add {
			rsl = append(rsl, transaction.Ident)
		}
	}
	return strings.Join(rsl, ", ")
}

type Transactions []Transaction

func JournalFromTable(table Table) Transactions {
	rsl := []Transaction{}
	for _, row := range table.RowList {
		if row.Section == "*" || row.DocLink == "" {
			continue
		}
		rsl = append(rsl, TransactionFromRow(row))
	}
	return rsl
}

func (t Transactions) Len() int {
	return len(t)
}
func (t Transactions) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
func (t Transactions) Less(i, j int) bool {
	return t[i].Ident < t[j].Ident
}

type Transaction struct {
	Unique           string
	Section          string
	Date             string
	Ident            string
	Path             string
	Description      string
	AccountDebit     string
	AccountCredit    string
	Amount           string // Always base currency.
	Currency         string
	AmountCurrency   string
	ExchangeCurrency string
	ExchangeRate     string
	Cc3              string // Cost center 3
	Cc3Des           string // Cost center 3 description

	// Cash basis accounting (EÜR) fields
	Income      string
	Expenses    string
	Account     string
	Category    string
	CategoryDes string
}

func TransactionFromRow(row Row) Transaction {
	return Transaction{
		Unique:           row.Unique,
		Section:          row.Section,
		Ident:            row.Doc,
		Date:             row.Date,
		Path:             row.DocLink,
		Description:      row.Description,
		AccountDebit:     row.AccountDebit,
		AccountCredit:    row.AccountCredit,
		Amount:           row.Amount,
		Currency:         row.Currency,
		AmountCurrency:   row.AmountCurrency,
		ExchangeCurrency: row.ExchangeCurrency,
		ExchangeRate:     row.ExchangeRate,
		Cc3:              row.Cc3,
		Cc3Des:           row.Cc3Des,

		// Cash basis accounting (EÜR) fields
		Income:      row.Income,
		Expenses:    row.Expenses,
		Account:     row.Account,
		Category:    row.Category,
		CategoryDes: row.CategoryDes,
	}
}

func (t Transaction) ParsedDate() (time.Time, error) {
	return time.Parse("2006-01-02", t.Date)
}

func (t Transaction) FmtDate() string {
	date, err := t.ParsedDate()
	if err != nil {
		return "<UNDEFINED>"
	}
	return date.Format(DATE_FORMAT)
}

func (t Transaction) FmtDescription() string {
	return removeExtraSpaces(t.Description)
}

func (t Transaction) GetAccountDebit(cashBasisAccounting bool) string {
	if cashBasisAccounting {
		return t.Account
	}
	return t.AccountDebit
}

func (t Transaction) GetAccountCredit(cashBasisAccounting bool) string {
	if cashBasisAccounting {
		return t.Category
	}
	return t.AccountCredit
}

func (t Transaction) GetAmount(cashBasisAccounting bool) string {
	if cashBasisAccounting {
		var rsl []string
		if t.Income != "" {
			rsl = append(rsl, t.Income)
		}
		if t.Expenses != "" {
			rsl = append(rsl, t.Expenses)
		}
		return strings.Join(rsl, "/")
	}
	return t.AccountDebit
}

// To represent accounts payable (AP) and receivable (AR) transactions within the cash
// basis accounting (EÜR) framework, I record an expense without specifying an
// account and category to cost centre 3 upon receipt of an invoice. These
// bookings are immaterial for the accounting and should therefore be
// clearly highlighted as such in the generated PDF.
//
// AP/AR auxiliary/immaterial transaction:
// - Account: Empty
// - Category: Empty
// - Cost centre 3: Set to supplier/customer
func (t Transaction) IsAPARAuxiliary(cashBasisAccounting bool) bool {
	return cashBasisAccounting && t.Account == "" && t.Category == "" && t.Cc3 != ""
}

func removeExtraSpaces(text string) string {
	var builder strings.Builder
	lastWasSpace := false

	for _, char := range text {
		if char == ' ' || char == '\t' || char == '\n' || char == '\r' {
			if !lastWasSpace {
				builder.WriteRune(' ')
				lastWasSpace = true
			}
			continue
		}
		builder.WriteRune(char)
		lastWasSpace = false
	}
	return strings.TrimSpace(builder.String())
}
