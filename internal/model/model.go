package model

import "time"

// SystemTransaction merepresentasikan transaksi internal sistem.
type SystemTransaction struct {
    TrxID           string
    Amount          int64  // asumsi: satuan Rupiah (tanpa desimal)
    Type            string // "DEBIT" atau "CREDIT"
    TransactionTime time.Time
}

// BankStatement merepresentasikan transaksi dari bank (per file/bank).
type BankStatement struct {
    UniqueIdentifier string
    Amount           int64 // bertanda: debit negatif, kredit positif
    Date             time.Time // hanya tanggal (time komponen diabaikan)
    BankName         string
}

// NormalizedRecord untuk matching per tanggal + tanda amount.
type NormalizedRecord struct {
    ID     string
    Date   time.Time // diseragamkan ke tanggal
    Amount int64     // signed
}

// MatchedPair hasil pasangan matched system vs bank.
type MatchedPair struct {
    SystemID     string
    BankID       string
    BankName     string
    Date         string
    SystemAmount int64
    BankAmount   int64
    Discrepancy  int64 // |SystemAmount - BankAmount|
}

// Result ringkasan dan detail rekonsiliasi.
type Result struct {
    Summary Summary `json:"summary"`
    Details Details `json:"details"`
}

type Summary struct {
    TotalProcessed     int   `json:"total_processed"`
    TotalMatched       int   `json:"total_matched"`
    TotalUnmatched     int   `json:"total_unmatched"`
    TotalDiscrepancies int64 `json:"total_discrepancies"`
}

type Details struct {
    Matched              []MatchedPair            `json:"matched"`
    UnmatchedSystem      []NormalizedRecord       `json:"unmatched_system"`
    UnmatchedBankByGroup map[string][]NormalizedRecord `json:"unmatched_bank_by_group"`
}