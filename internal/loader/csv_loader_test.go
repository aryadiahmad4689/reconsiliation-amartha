package loader

import (
    "os"
    "path/filepath"
    "testing"
    "time"
)

func writeTempFile(t *testing.T, dir, name, content string) string {
    t.Helper()
    p := filepath.Join(dir, name)
    if err := os.WriteFile(p, []byte(content), 0644); err != nil {
        t.Fatalf("write file: %v", err)
    }
    return p
}

func TestLoadSystemCSV_OK(t *testing.T) {
    dir := t.TempDir()
    content := "trxID,amount,type,transactionTime\n" +
        "TRX-1,250000,CREDIT,2025-06-01T12:34:56Z\n" +
        "TRX-2,-125000,DEBIT,2025-06-01T15:00:00Z\n" +
        "TRX-3,\"495,000\",CREDIT,2025-06-02T08:00:00Z\n"
    p := writeTempFile(t, dir, "system.csv", content)

    got, err := LoadSystemCSV(p)
    if err != nil {
        t.Fatalf("LoadSystemCSV error: %v", err)
    }
    if len(got) != 3 {
        t.Fatalf("len(got)=%d", len(got))
    }
    if got[0].TrxID != "TRX-1" || got[0].Amount != 250000 || got[0].Type != "CREDIT" || !got[0].TransactionTime.Equal(time.Date(2025, 6, 1, 12, 34, 56, 0, time.UTC)) {
        t.Fatalf("unexpected first row: %+v", got[0])
    }
    if got[1].TrxID != "TRX-2" || got[1].Amount != -125000 || got[1].Type != "DEBIT" || !got[1].TransactionTime.Equal(time.Date(2025, 6, 1, 15, 0, 0, 0, time.UTC)) {
        t.Fatalf("unexpected second row: %+v", got[1])
    }
    if got[2].TrxID != "TRX-3" || got[2].Amount != 495000 || got[2].Type != "CREDIT" || !got[2].TransactionTime.Equal(time.Date(2025, 6, 2, 8, 0, 0, 0, time.UTC)) {
        t.Fatalf("unexpected third row: %+v", got[2])
    }
}

func TestLoadSystemCSV_InvalidAmount(t *testing.T) {
    dir := t.TempDir()
    content := "trxID,amount,type,transactionTime\n" +
        "TRX-1,abc,CREDIT,2025-06-01T12:34:56Z\n"
    p := writeTempFile(t, dir, "system_bad_amount.csv", content)
    if _, err := LoadSystemCSV(p); err == nil {
        t.Fatalf("expected error for invalid amount")
    }
}

func TestLoadSystemCSV_InvalidTime(t *testing.T) {
    dir := t.TempDir()
    content := "trxID,amount,type,transactionTime\n" +
        "TRX-1,1000,CREDIT,2025/06/01 12:34:56\n"
    p := writeTempFile(t, dir, "system_bad_time.csv", content)
    if _, err := LoadSystemCSV(p); err == nil {
        t.Fatalf("expected error for invalid time")
    }
}

func TestLoadBankCSV_OK(t *testing.T) {
    dir := t.TempDir()
    content := "unique_identifier,amount,date\n" +
        "BA-1,250000,2025-06-01\n" +
        "BB-2,\"-75,000\",2025-06-03\n"
    p := writeTempFile(t, dir, "bank.csv", content)

    got, err := LoadBankCSV(p, "bankA")
    if err != nil {
        t.Fatalf("LoadBankCSV error: %v", err)
    }
    if len(got) != 2 {
        t.Fatalf("len(got)=%d", len(got))
    }
    if got[0].UniqueIdentifier != "BA-1" || got[0].Amount != 250000 || !got[0].Date.Equal(time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)) || got[0].BankName != "bankA" {
        t.Fatalf("unexpected first row: %+v", got[0])
    }
    if got[1].UniqueIdentifier != "BB-2" || got[1].Amount != -75000 || !got[1].Date.Equal(time.Date(2025, 6, 3, 0, 0, 0, 0, time.UTC)) || got[1].BankName != "bankA" {
        t.Fatalf("unexpected second row: %+v", got[1])
    }
}

func TestLoadBankCSV_InvalidAmount(t *testing.T) {
    dir := t.TempDir()
    content := "unique_identifier,amount,date\n" +
        "BA-1,xyz,2025-06-01\n"
    p := writeTempFile(t, dir, "bank_bad_amount.csv", content)
    if _, err := LoadBankCSV(p, "bankA"); err == nil {
        t.Fatalf("expected error for invalid amount")
    }
}

func TestLoadBankCSV_InvalidDate(t *testing.T) {
    dir := t.TempDir()
    content := "unique_identifier,amount,date\n" +
        "BA-1,1000,01-06-2025\n"
    p := writeTempFile(t, dir, "bank_bad_date.csv", content)
    if _, err := LoadBankCSV(p, "bankA"); err == nil {
        t.Fatalf("expected error for invalid date")
    }
}

func TestParseAmount(t *testing.T) {
    cases := []struct {
        in  string
        out int64
        ok  bool
    }{
        {"0", 0, true},
        {"123", 123, true},
        {"-456", -456, true},
        {"1,234", 1234, true},
        {"-250,000", -250000, true},
        {"abc", 0, false},
    }
    for _, c := range cases {
        v, err := parseAmount(c.in)
        if c.ok {
            if err != nil || v != c.out {
                t.Fatalf("parseAmount(%q) => %v,%v", c.in, v, err)
            }
        } else {
            if err == nil {
                t.Fatalf("expected error for %q", c.in)
            }
        }
    }
}