package main

import "strings"

// Everything with the same linked document. Filepath is map key.
type Dossier struct {
	JournalEntries     Entries
	AccountingFilePath string
	CompanyName        string
	Street             string
	ZIPCode            string
	Place              string
	LastSaved          string
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
		AccountingFilePath: fileInfoTable.GuardedValueById("Dateiname"),
		CompanyName:        fileInfoTable.GuardedValueById("Firma"),
		Street:             fileInfoTable.GuardedValueById("Adresse1"),
		ZIPCode:            fileInfoTable.GuardedValueById("Postleitzahl"),
		Place:              fileInfoTable.GuardedValueById("Ort"),
		LastSaved:          fileInfoTable.GuardedValueById("ZeitLetzteSpeicherung"),
	}, nil
}

type Entries map[string]Document

func EntriesFromJournal(journal Journal) Entries {
	rsl := map[string]Document{}
	for _, transaction := range journal {
		if _, exists := rsl[transaction.Path]; !exists {
			rsl[transaction.Path] = Document{}
		}
		if _, exists := rsl[transaction.Path][transaction.Ident]; !exists {
			rsl[transaction.Path][transaction.Ident] = []Transaction{}
		}
		rsl[transaction.Path][transaction.Ident] = append(rsl[transaction.Path][transaction.Ident], transaction)
	}
	return rsl
}

// All docs of one doc-ident. Doc-ident is map key.
type Document map[string][]Transaction

func (d Document) IdentStringList() string {
	rsl := []string{}
	for ident, _ := range d {
		rsl = append(rsl, ident)
	}
	return strings.Join(rsl, " — ")
}

type Journal []Transaction

func JournalFromTable(table Table) Journal {
	rsl := []Transaction{}
	for _, row := range table.RowList {
		if row.Section == "*" || row.DocLink == "" {
			continue
		}
		rsl = append(rsl, TransactionFromRow(row))
	}
	return rsl
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
		AmountCurrency:   row.AmountCurrency,
		ExchangeCurrency: row.ExchangeCurrency,
		ExchangeRate:     row.ExchangeRate,
	}
}
