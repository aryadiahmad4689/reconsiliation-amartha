package reconcile

import (
	"sort"
	"strings"
	"time"

	"amartha/internal/loader"
	"amartha/internal/model"
)

const discrepancyTolerance int64 = 5000

// bankRec adalah representasi record bank yang disertai nama bank untuk pelaporan.
type bankRec struct {
	model.NormalizedRecord
	BankName string
}

// Reconcile melakukan rekonsiliasi antara transaksi sistem dan bank dalam rentang tanggal.
// Strategi matching: per tanggal dan tanda amount; pasangan dibentuk dengan
// mengurutkan amount dan dipasangkan berurutan (minimalkan total selisih absolut).
func Reconcile(sys []model.SystemTransaction, banks map[string][]loader.BankStatement, start, end time.Time) (model.Result, error) {
	// Filter dan normalisasi sistem.
	var sysPos, sysNeg []model.NormalizedRecord // pos: credit, neg: debit
	for _, s := range sys {
		d := s.TransactionTime.In(time.UTC)
		dateOnly := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
		if dateOnly.Before(start) || dateOnly.After(end) {
			continue
		}
		signed := s.Amount
		if strings.EqualFold(s.Type, "DEBIT") {
			signed = -signed
		}
		rec := model.NormalizedRecord{ID: s.TrxID, Date: dateOnly, Amount: signed}
		if signed >= 0 {
			sysPos = append(sysPos, rec)
		} else {
			sysNeg = append(sysNeg, rec)
		}
	}

	// Filter dan normalisasi bank.
	var bankPos, bankNeg []bankRec
	for bankName, list := range banks {
		for _, b := range list {
			d := time.Date(b.Date.Year(), b.Date.Month(), b.Date.Day(), 0, 0, 0, 0, time.UTC)
			if d.Before(start) || d.After(end) {
				continue
			}
			br := bankRec{NormalizedRecord: model.NormalizedRecord{ID: b.UniqueIdentifier, Date: d, Amount: b.Amount}, BankName: bankName}
			if b.Amount >= 0 {
				bankPos = append(bankPos, br)
			} else {
				bankNeg = append(bankNeg, br)
			}
		}
	}

	matched := []model.MatchedPair{}
	unmatchedSys := []model.NormalizedRecord{}
	unmatchedBankByGroup := map[string][]model.NormalizedRecord{}

	// Proses per tanda dan per tanggal.
	matchedPos, umSysPos, umBankPos := matchByDateAndAmount(sysPos, bankPos)
	matchedNeg, umSysNeg, umBankNeg := matchByDateAndAmount(sysNeg, bankNeg)

	matched = append(matched, matchedPos...)
	matched = append(matched, matchedNeg...)
	unmatchedSys = append(unmatchedSys, umSysPos...)
	unmatchedSys = append(unmatchedSys, umSysNeg...)
	for bankName, recs := range umBankPos {
		unmatchedBankByGroup[bankName] = append(unmatchedBankByGroup[bankName], recs...)
	}
	for bankName, recs := range umBankNeg {
		unmatchedBankByGroup[bankName] = append(unmatchedBankByGroup[bankName], recs...)
	}

	// Ringkasan.
	totalProcessed := len(sysPos) + len(sysNeg) + len(bankPos) + len(bankNeg)
	var totalDiscrepancies int64
	for _, m := range matched {
		totalDiscrepancies += abs64(m.SystemAmount - m.BankAmount)
	}
	totalUnmatched := len(unmatchedSys)
	for _, recs := range unmatchedBankByGroup {
		totalUnmatched += len(recs)
	}

	return model.Result{
		Summary: model.Summary{
			TotalProcessed:     totalProcessed,
			TotalMatched:       len(matched),
			TotalUnmatched:     totalUnmatched,
			TotalDiscrepancies: totalDiscrepancies,
		},
		Details: model.Details{
			Matched:              matched,
			UnmatchedSystem:      unmatchedSys,
			UnmatchedBankByGroup: unmatchedBankByGroup,
		},
	}, nil
}

// matchByDateAndAmount melakukan pairing per tanggal yang sama, dengan mengurutkan amount
// untuk meminimalkan total selisih absolut. Bank rec menyimpan nama bank untuk pelaporan.
func matchByDateAndAmount(sys []model.NormalizedRecord, bank []bankRec) (
	[]model.MatchedPair,
	[]model.NormalizedRecord,
	map[string][]model.NormalizedRecord,
) {
	sysByDate := groupByDateSys(sys)
	bankByDate := groupByDateBank(bank)
	ds := collectSortedDates(sysByDate, bankByDate)

	matched := []model.MatchedPair{}
	unmatchedSys := []model.NormalizedRecord{}
	unmatchedBank := map[string][]model.NormalizedRecord{}

	for _, d := range ds {
		m, umS, umB := pairForDate(d, sysByDate[d], bankByDate[d])
		matched = append(matched, m...)
		unmatchedSys = append(unmatchedSys, umS...)
		for k, v := range umB {
			unmatchedBank[k] = append(unmatchedBank[k], v...)
		}
	}

	return matched, unmatchedSys, unmatchedBank
}

// abs64 mengembalikan nilai absolut dari bilangan bertanda int64.
func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

// groupByDateSys mengelompokkan record sistem berdasarkan tanggalnya.
func groupByDateSys(sys []model.NormalizedRecord) map[time.Time][]model.NormalizedRecord {
	out := map[time.Time][]model.NormalizedRecord{}
	for _, s := range sys {
		out[s.Date] = append(out[s.Date], s)
	}
	return out
}

// groupByDateBank mengelompokkan record bank berdasarkan tanggalnya.
func groupByDateBank(bank []bankRec) map[time.Time][]bankRec {
	out := map[time.Time][]bankRec{}
	for _, b := range bank {
		out[b.Date] = append(out[b.Date], b)
	}
	return out
}

// collectSortedDates mengambil union tanggal dari sistem dan bank,
// lalu mengembalikannya sebagai slice yang diurutkan kronologis.
func collectSortedDates(sysByDate map[time.Time][]model.NormalizedRecord, bankByDate map[time.Time][]bankRec) []time.Time {
	set := make(map[time.Time]struct{})
	for d := range sysByDate {
		set[d] = struct{}{}
	}
	for d := range bankByDate {
		set[d] = struct{}{}
	}
	ds := make([]time.Time, 0, len(set))
	for d := range set {
		ds = append(ds, d)
	}
	sort.Slice(ds, func(i, j int) bool { return ds[i].Before(ds[j]) })
	return ds
}

// pairForDate mencocokkan record sistem dan bank untuk satu tanggal tertentu.
// Daftar diurutkan berdasarkan amount, lalu dipasangkan dengan two-pointer
// menggunakan toleransi selisih untuk menentukan pasangan matched dan elemen unmatched.
func pairForDate(d time.Time, sList []model.NormalizedRecord, bList []bankRec) (
	[]model.MatchedPair,
	[]model.NormalizedRecord,
	map[string][]model.NormalizedRecord,
) {
	matched := []model.MatchedPair{}
	unmatchedSys := []model.NormalizedRecord{}
	unmatchedBank := map[string][]model.NormalizedRecord{}

	// mengurutkan amount
	sort.Slice(sList, func(i, j int) bool { return sList[i].Amount < sList[j].Amount })
	sort.Slice(bList, func(i, j int) bool { return bList[i].Amount < bList[j].Amount })

	i, j := 0, 0
	for i < len(sList) && j < len(bList) {
		s := sList[i]
		b := bList[j]
		diff := abs64(s.Amount - b.Amount)
		if diff <= discrepancyTolerance {
			matched = append(matched, model.MatchedPair{
				SystemID:     s.ID,
				BankID:       b.ID,
				BankName:     b.BankName,
				Date:         d.Format("2006-01-02"),
				SystemAmount: s.Amount,
				BankAmount:   b.Amount,
				Discrepancy:  diff,
			})
			i++
			j++
		} else if s.Amount < b.Amount {
			unmatchedSys = append(unmatchedSys, s)
			i++
		} else {
			unmatchedBank[b.BankName] = append(unmatchedBank[b.BankName], b.NormalizedRecord)
			j++
		}
	}
	if i < len(sList) {
		unmatchedSys = append(unmatchedSys, sList[i:]...)
	}
	if j < len(bList) {
		for ; j < len(bList); j++ {
			b := bList[j]
			unmatchedBank[b.BankName] = append(unmatchedBank[b.BankName], b.NormalizedRecord)
		}
	}

	return matched, unmatchedSys, unmatchedBank
}
