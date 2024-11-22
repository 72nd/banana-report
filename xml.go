package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

type AC2 struct {
	XMLName            xml.Name           `xml:"AC2"`
	Version            string             `xml:"version,attr"`
	DocumentProperties DocumentProperties `xml:"DocumentProperties"`
	Styles             Styles             `xml:"Styles"`
	Tables             []Table            `xml:"Table"`
}

func AC2FromFile(path string) (*AC2, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	raw, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var rsl AC2
	err = xml.Unmarshal(raw, &rsl)
	if err != nil {
		return nil, err
	}
	return &rsl, nil
}

func (a AC2) ToJSON(path string) error {
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (a AC2) TableById(id string) (*Table, error) {
	for _, table := range a.Tables {
		if table.ID == id {
			return &table, nil
		}
	}
	return nil, fmt.Errorf("no Table for ID '%s' found", id)
}

type DocumentProperties struct {
	Version        string `xml:"Version"`
	VersionCreated string `xml:"VersionCreated"`
}

type Styles struct {
	Styles []Style `xml:"Style"`
}

type Style struct {
	ID   string `xml:"ID,attr"`
	Size string `xml:"Size"`
	Bold string `xml:"Bold,omitempty"`
}

type Table struct {
	ID         string  `xml:"ID,attr"`
	Name       string  `xml:"Name"`
	IDNumber   string  `xml:"IDNumber"`
	Header     string  `xml:"Header"`
	PeriodFreq string  `xml:"PeriodFreq,omitempty"`
	FieldList  []Field `xml:"FieldList>Field"`
	RowList    []Row   `xml:"RowList>Row"`
}

func (t Table) GuardedValueById(id string) string {
	rsl, err := t.ValueById(id)
	if err != nil {
		fmt.Println(err)
		return "<UNDEFINED>"
	}
	return rsl
}

func (t Table) ValueById(id string) (string, error) {
	for _, row := range t.RowList {
		if row.NodeID != id || row.Value == "" {
			continue
		}
		return row.Value, nil
	}
	return "", fmt.Errorf("couldn't find value for id '%s'", id)
}

type Field struct {
	ID              string `xml:"ID,attr"`
	Sequence        string `xml:"Sequence"`
	Number          string `xml:"Number"`
	Name            string `xml:"Name"`
	Datatype        string `xml:"Datatype"`
	MaxLength       string `xml:"MaxLength,omitempty"`
	System          string `xml:"System"`
	SystemProtected string `xml:"SystemProtected,omitempty"`
	Protected       string `xml:"Protected,omitempty"`
	Header1         string `xml:"Header1"`
	Description     string `xml:"Description"`
	Width           string `xml:"Width"`
}

type Row struct {
	ID                         string `xml:"ID,attr"`
	NodeID                     string `xml:"Id"`
	Unique                     string `xml:"Unique"`
	Date                       string `xml:"Date,omitempty"`
	BaseRow                    string `xml:"BaseRow,omitempty"`
	Style                      string `xml:"Style,omitempty"`
	Section                    string `xml:"Section,omitempty"`
	Description                string `xml:"Description,omitempty"`
	Account                    string `xml:"Account,omitempty"`
	AccountDebit               string `xml:"AccountDebit,omitempty"`
	AccountDebitDes            string `xml:"AccountDebitDes,omitempty"`
	AccountCredit              string `xml:"AccountCredit,omitempty"`
	AccountCreditDes           string `xml:"AccountCreditDes,omitempty"`
	AmountCurrency             string `xml:"AmountCurrency,omitempty"`
	ExchangeCurrency           string `xml:"ExchangeCurrency,omitempty"`
	ExchangeRate               string `xml:"ExchangeRate,omitempty"`
	ExchangeMultiplier         string `xml:"ExchangeMultiplier,omitempty"`
	Amount                     string `xml:"Amount,omitempty"`
	Balance                    string `xml:"Balance,omitempty"`
	Currency                   string `xml:"Currency,omitempty"`
	OpeningCurrency            string `xml:"OpeningCurrency,omitempty"`
	BalanceCalculatedCurrency2 string `xml:"BalanceCalculatedCurrency2,omitempty"`
	Doc                        string `xml:"Doc,omitempty"`
	DocInvoice                 string `xml:"DocInvoice,omitempty"`
	DocLink                    string `xml:"DocLink,omitempty"`
	Value                      string `xml:"Value,omitempty"`
}
