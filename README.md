# Reconciliation Service (Go)

Layanan untuk merekonsiliasi transaksi sistem dengan bank statements dalam rentang tanggal.

## Struktur Folder

```
amartha/
├─ cmd/
│  └─ reconcile/
│     └─ main.go            # CLI entrypoint
├─ internal/
│  ├─ loader/
│  │  └─ csv_loader.go      # Parser CSV sistem & bank
│  ├─ model/
│  │  └─ model.go           # Definisi struct domain & hasil
│  └─ reconcile/
│     └─ engine.go          # Algoritma rekonsiliasi
├─ testdata/                # Contoh input CSV
│  ├─ system_transactions.csv
│  ├─ bankA.csv
│  └─ bankB.csv
└─ go.mod
```

## Asumsi Desain

- Amount dalam satuan Rupiah integer (`int64`), tanpa desimal.
- `type` sistem: `CREDIT` (positif), `DEBIT` (negatif). Bank amount sudah bertanda.
- Matching dilakukan per tanggal & tanda amount; untuk meminimalkan total selisih, kedua sisi diurutkan berdasarkan amount dan dipasangkan dua-pointer.
- Discrepancy adalah `|amount_system - amount_bank|` pada pasangan matched. Toleransi selisih default: `5000`.
- Nama bank diambil dari nama file CSV bank (tanpa ekstensi) untuk pelaporan per bank.

## Cara Menjalankan

1. Pastikan Go 1.21+.
2. Siapkan input CSV. Secara default, program menggunakan file di folder `testdata/` (lihat konfigurasi di `cmd/reconcile/main.go`).
   - Ubah isi `main.go` jika ingin mengganti path atau tanggal.
3. Jalankan (mode CLI):

```
go run ./cmd/reconcile \
  --system ./testdata/system_transactions.csv \
  --bank ./testdata/bankA.csv \
  --bank ./testdata/bankB.csv \
  --start 2025-06-01 \
  --end 2025-06-03
```

Output berupa JSON ringkasan dan detail hasil rekonsiliasi.

## Format CSV

System (`system_transactions.csv`):

```
trxID,amount,type,transactionTime
TRX-1001,250000,CREDIT,2025-06-01T10:15:00Z
TRX-1002,125000,DEBIT,2025-06-01T12:00:00Z
TRX-1003,500000,CREDIT,2025-06-02T09:00:00Z
TRX-1004,180000,CREDIT,2025-06-02T14:21:00Z
TRX-1005,75000,DEBIT,2025-06-03T08:00:00Z
TRX-1006,42000,CREDIT,2025-06-03T17:00:00Z
```

Bank A (`bankA.csv`):

```
unique_identifier,amount,date
BA-7781,250000,2025-06-01
BA-7782,-125000,2025-06-01
BA-7783,495000,2025-06-02
BA-7784,180000,2025-06-02
```

Bank B (`bankB.csv`):

```
unique_identifier,amount,date
BB-3001,-75000,2025-06-03
BB-3002,100000,2025-06-03
```

## Testing

Tambahkan unit test di `internal/reconcile` untuk memverifikasi perhitungan matched, unmatched, dan discrepancy. Contoh test dapat menggunakan `testdata` yang disediakan.

Cara menjalankan unit test:

```
# Jalankan semua paket
go test ./... -count=1

# Jalankan hanya paket rekonsiliasi dengan output detail
go test amartha/internal/reconcile -v -count=1

# Jalankan test case tertentu saja
go test -run ^TestReconcileBasic$ amartha/internal/reconcile -v -count=1

# Tambahkan timeout bila perlu
go test amartha/internal/reconcile -v -count=1 -timeout 30s
```
