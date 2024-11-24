package main

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Everything with the same linked document. Filepath is map key.
type Dossier struct {
	JournalEntries    Documents
	AccountingDirPath string
	CompanyName       string
	Street            string
	ZIPCode           string
	Place             string
	LastSaved         string
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
		JournalEntries:    entries,
		AccountingDirPath: filepath.Dir(fileInfoTable.GuardedValueById("Dateiname")),
		CompanyName:       fileInfoTable.GuardedValueById("Firma"),
		Street:            fileInfoTable.GuardedValueById("Adresse1"),
		ZIPCode:           fileInfoTable.GuardedValueById("Postleitzahl"),
		Place:             fileInfoTable.GuardedValueById("Ort"),
		LastSaved:         fileInfoTable.GuardedValueById("ZeitLetzteSpeicherung"),
	}, nil
}

func (d Dossier) ResolveRelativePath(path string) (string, error) {
	fullPath := filepath.Join(d.AccountingDirPath, path)
	return filepath.Abs(fullPath)
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
	Amount           string
	Currency         string
	AmountCurrency   string
	ExchangeCurrency string
	ExchangeRate     string
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
	return date.Format("02.01.06")
}

func (t Transaction) FmtDescription() string {
	return removeExtraSpaces(t.Description)
}

func (t Transaction) FmtAmount() string {
	return fmt.Sprintf("%s CHF", t.Amount)
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
