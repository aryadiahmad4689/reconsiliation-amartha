package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"amartha/internal/loader"
	"amartha/internal/model"
	"amartha/internal/reconcile"
)

func main() {
	sysPath, bankPaths, start, end := parseArgs()
	sysTxs := mustLoadSystemCSV(sysPath)
	bankData := mustLoadBanks(bankPaths)
	res, err := reconcile.Reconcile(sysTxs, bankData, start, end)
	if err != nil {
		log.Fatalf("reconciliation error: %v", err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(res); err != nil {
		log.Fatalf("failed to encode result: %v", err)
	}
}

func parseArgs() (string, []string, time.Time, time.Time) {
	systemPath := flag.String("system", "", "Path to system transactions CSV (required)")
	var bankPaths multiFlag
	flag.Var(&bankPaths, "bank", "Path to bank statement CSV (repeatable)")
	startStr := flag.String("start", "", "Start date YYYY-MM-DD (inclusive)")
	endStr := flag.String("end", "", "End date YYYY-MM-DD (inclusive)")
	flag.Parse()
	if *systemPath == "" || len(bankPaths) == 0 || *startStr == "" || *endStr == "" {
		flag.Usage()
		os.Exit(2)
	}
	start, err := time.Parse("2006-01-02", *startStr)
	if err != nil {
		log.Fatalf("invalid start date: %v", err)
	}
	end, err := time.Parse("2006-01-02", *endStr)
	if err != nil {
		log.Fatalf("invalid end date: %v", err)
	}
	if end.Before(start) {
		log.Fatalf("end date must be on or after start date")
	}
	return *systemPath, []string(bankPaths), start, end
}

func mustLoadSystemCSV(p string) []model.SystemTransaction {
	txs, err := loader.LoadSystemCSV(p)
	if err != nil {
		log.Fatalf("failed to load system CSV: %v", err)
	}
	return txs
}

func mustLoadBanks(paths []string) map[string][]loader.BankStatement {
	out := make(map[string][]loader.BankStatement)
	for _, p := range paths {
		name := bankNameFromPath(p)
		bs, err := loader.LoadBankCSV(p, name)
		if err != nil {
			log.Fatalf("failed to load bank CSV %s: %v", p, err)
		}
		out[name] = bs
	}
	return out
}

func bankNameFromPath(p string) string {
	base := filepath.Base(p)
	ext := filepath.Ext(base)
	return base[:len(base)-len(ext)]
}

type multiFlag []string

func (m *multiFlag) String() string         { return fmt.Sprint([]string(*m)) }
func (m *multiFlag) Set(value string) error { *m = append(*m, value); return nil }
