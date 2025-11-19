package reconcile

import (
    "testing"
    "time"

    "amartha/internal/loader"
    "amartha/internal/model"
)

func TestReconcileBasic(t *testing.T) {
    sys := []model.SystemTransaction{
        {TrxID: "TRX-1001", Amount: 250000, Type: "CREDIT", TransactionTime: mustRFC3339("2025-06-01T10:15:00Z")},
        {TrxID: "TRX-1002", Amount: 125000, Type: "DEBIT", TransactionTime: mustRFC3339("2025-06-01T12:00:00Z")},
        {TrxID: "TRX-1003", Amount: 500000, Type: "CREDIT", TransactionTime: mustRFC3339("2025-06-02T09:00:00Z")},
        {TrxID: "TRX-1004", Amount: 180000, Type: "CREDIT", TransactionTime: mustRFC3339("2025-06-02T14:21:00Z")},
        {TrxID: "TRX-1005", Amount: 75000, Type: "DEBIT", TransactionTime: mustRFC3339("2025-06-03T08:00:00Z")},
        {TrxID: "TRX-1006", Amount: 42000, Type: "CREDIT", TransactionTime: mustRFC3339("2025-06-03T17:00:00Z")},
    }
    banks := map[string][]loader.BankStatement{
        "bankA": {
            {UniqueIdentifier: "BA-7781", Amount: 250000, Date: mustDate("2025-06-01"), BankName: "bankA"},
            {UniqueIdentifier: "BA-7782", Amount: -125000, Date: mustDate("2025-06-01"), BankName: "bankA"},
            {UniqueIdentifier: "BA-7783", Amount: 495000, Date: mustDate("2025-06-02"), BankName: "bankA"},
            {UniqueIdentifier: "BA-7784", Amount: 180000, Date: mustDate("2025-06-02"), BankName: "bankA"},
        },
        "bankB": {
            {UniqueIdentifier: "BB-3001", Amount: -75000, Date: mustDate("2025-06-03"), BankName: "bankB"},
            {UniqueIdentifier: "BB-3002", Amount: 100000, Date: mustDate("2025-06-03"), BankName: "bankB"},
        },
    }

    start := mustDate("2025-06-01")
    end := mustDate("2025-06-03")

    res, err := Reconcile(sys, banks, start, end)
    if err != nil { t.Fatalf("error: %v", err) }

    if res.Summary.TotalProcessed != 12 {
        t.Fatalf("expected total processed 12, got %d", res.Summary.TotalProcessed)
    }
    if res.Summary.TotalMatched != 5 {
        t.Fatalf("expected total matched 5, got %d", res.Summary.TotalMatched)
    }
    if res.Summary.TotalDiscrepancies != 5000 {
        t.Fatalf("expected discrepancies 5000, got %d", res.Summary.TotalDiscrepancies)
    }
    // unmatched: system (TRX-1006), bank (BB-3002)
    if res.Summary.TotalUnmatched != 2 {
        t.Fatalf("expected total unmatched 2, got %d", res.Summary.TotalUnmatched)
    }
}

func mustRFC3339(s string) time.Time {
    t, err := time.Parse(time.RFC3339, s)
    if err != nil { panic(err) }
    return t
}
func mustDate(s string) time.Time {
    t, err := time.Parse("2006-01-02", s)
    if err != nil { panic(err) }
    return t
}