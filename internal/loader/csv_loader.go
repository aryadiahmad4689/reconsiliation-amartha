package loader

import (
    "encoding/csv"
    "fmt"
    "io"
    "os"
    "strconv"
    "time"

    "amartha/internal/model"
)

// LoadSystemCSV membaca CSV transaksi sistem.
// Format header: trxID,amount,type,transactionTime
func LoadSystemCSV(path string) ([]model.SystemTransaction, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    r := csv.NewReader(f)
    r.TrimLeadingSpace = true

    // baca header
    if _, err := r.Read(); err != nil {
        return nil, err
    }

    var out []model.SystemTransaction
    for {
        rec, err := r.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, err
        }
        if len(rec) < 4 {
            return nil, fmt.Errorf("invalid system csv row: %v", rec)
        }
        amt, err := parseAmount(rec[1])
        if err != nil {
            return nil, fmt.Errorf("invalid amount %q: %w", rec[1], err)
        }
        t, err := time.Parse(time.RFC3339, rec[3])
        if err != nil {
            return nil, fmt.Errorf("invalid transactionTime %q: %w", rec[3], err)
        }
        out = append(out, model.SystemTransaction{
            TrxID:           rec[0],
            Amount:          amt,
            Type:            rec[2],
            TransactionTime: t,
        })
    }
    return out, nil
}

// LoadBankCSV membaca CSV bank statement.
// Format header: unique_identifier,amount,date
func LoadBankCSV(path string, bankName string) ([]BankStatement, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    r := csv.NewReader(f)
    r.TrimLeadingSpace = true

    if _, err := r.Read(); err != nil {
        return nil, err
    }

    var out []BankStatement
    for {
        rec, err := r.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, err
        }
        if len(rec) < 3 {
            return nil, fmt.Errorf("invalid bank csv row: %v", rec)
        }
        amt, err := parseAmount(rec[1])
        if err != nil {
            return nil, fmt.Errorf("invalid amount %q: %w", rec[1], err)
        }
        d, err := time.Parse("2006-01-02", rec[2])
        if err != nil {
            return nil, fmt.Errorf("invalid date %q: %w", rec[2], err)
        }
        out = append(out, BankStatement{
            UniqueIdentifier: rec[0],
            Amount:           amt,
            Date:             d,
            BankName:         bankName,
        })
    }
    return out, nil
}

// BankStatement adalah versi loader untuk menyertakan nama bank.
type BankStatement struct {
    UniqueIdentifier string
    Amount           int64
    Date             time.Time
    BankName         string
}

func parseAmount(s string) (int64, error) {
    // asumsi Rupiah tanpa desimal. Mendukung tanda +/-. Mengabaikan koma pemisah ribuan.
    // Hapus koma jika ada.
    clean := ""
    for _, ch := range s {
        if ch == ',' {
            continue
        }
        clean += string(ch)
    }
    v, err := strconv.ParseInt(clean, 10, 64)
    if err != nil {
        return 0, err
    }
    return v, nil
}