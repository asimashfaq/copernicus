package mining

import (
	"github.com/copernet/copernicus/conf"
	"github.com/copernet/copernicus/log"
	"github.com/copernet/copernicus/model/mempool"
	"github.com/copernet/copernicus/util"
	"github.com/google/btree"
)

type sortType int

const (
	sortByFee sortType = 1 << iota
	sortByFeeRate
)

const defaultSortStrategy = sortByFeeRate

var strategy sortType

var strategies = map[string]sortType{
	"ancestorfee":     sortByFee,
	"ancestorfeerate": sortByFeeRate,
}

// EntryFeeSort TxEntry sorted by feeWithAncestors
type EntryFeeSort mempool.TxEntry

func (e EntryFeeSort) Less(than btree.Item) bool {
	t := than.(EntryFeeSort)
	if e.SumTxFeeWithAncestors == t.SumTxFeeWithAncestors {
		eHash := e.Tx.GetHash()
		tHash := t.Tx.GetHash()
		return eHash.Cmp(&tHash) > 0
	}
	return e.SumTxFeeWithAncestors < than.(EntryFeeSort).SumTxFeeWithAncestors
}

func sortedByFeeWithAncestors() *btree.BTree {
	b := btree.New(32)
	mpool := mempool.GetInstance()

	for _, txEntry := range mpool.GetAllTxEntry() {
		b.ReplaceOrInsert(EntryFeeSort(*txEntry))
	}

	return b
}

// EntryAncestorFeeRateSort TxEntry sorted by feeRateWithAncestors
type EntryAncestorFeeRateSort mempool.TxEntry

func (r EntryAncestorFeeRateSort) Less(than btree.Item) bool {
	t := than.(EntryAncestorFeeRateSort)
	b1 := util.NewFeeRateWithSize((r).SumTxFeeWithAncestors, r.SumTxSizeWitAncestors).SataoshisPerK
	b2 := util.NewFeeRateWithSize(t.SumTxFeeWithAncestors, t.SumTxSizeWitAncestors).SataoshisPerK
	if b1 == b2 {
		rHash := r.Tx.GetHash()
		tHash := t.Tx.GetHash()
		return rHash.Cmp(&tHash) > 0
	}
	return b1 < b2
}

func sortedByFeeRateWithAncestors() *btree.BTree {
	b := btree.New(32)
	mpool := mempool.GetInstance()

	for _, txEntry := range mpool.GetAllTxEntry() {
		b.ReplaceOrInsert(EntryAncestorFeeRateSort(*txEntry))
	}

	return b
}

func getStrategy() *sortType {
	if strategy != 0 {
		return &strategy
	}

	sortParam := conf.Cfg.Mining.Strategy
	ret, ok := strategies[sortParam]
	if !ok {
		log.Error("the specified strategy< %s > is not exist, so use default strategy< %s >", sortParam, defaultSortStrategy)
		strategy = defaultSortStrategy
	}
	strategy = ret
	return &strategy
}
