package indexer

import (
	"fmt"
	"time"

	common "github.com/arcology-network/common-lib/common"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/indexer"
	"github.com/arcology-network/concurrenturl/interfaces"
)

type Arbitrator struct{}

func (this *Arbitrator) Detect(newTrans []interfaces.Univalue) []*Conflict {
	if len(newTrans) == 0 {
		return []*Conflict{}
	}

	t0 := time.Now()
	indexer.Univalues(newTrans).Sort()
	fmt.Println("Sort: ", time.Since(t0))

	ranges := common.FindRange(newTrans, func(lhv, rhv interfaces.Univalue) bool { return *lhv.GetPath() == *rhv.GetPath() })

	conflicts := []*Conflict{}
	for i := 0; i < len(ranges)-1; i++ {
		if ranges[i]+1 == ranges[i+1] {
			continue // Only one entry
		}

		offset := int(1)
		if newTrans[ranges[i]].Writes() == 0 {
			if newTrans[ranges[i]].IsConcurrentWritable() { // Delta write only
				offset = common.LocateFirstIf(newTrans[ranges[i]+1:ranges[i+1]], func(v interfaces.Univalue) bool { return !v.IsConcurrentWritable() })
			} else { // Read only
				offset = common.LocateFirstIf(newTrans[ranges[i]+1:ranges[i+1]], func(v interfaces.Univalue) bool { return v.Writes() > 0 || v.DeltaWrites() > 0 })
			}
			offset = common.IfThen(offset < 0, ranges[i+1]-ranges[i], offset+1) // offset == -1 means no conflict found
		}

		if ranges[i]+offset == ranges[i+1] {
			continue
		}

		conflictTxs := []uint32{}
		common.Foreach(newTrans[ranges[i]+offset:ranges[i+1]], func(v *interfaces.Univalue) { conflictTxs = append(conflictTxs, (*v).GetTx()) })
		conflicts = append(conflicts,
			&Conflict{
				key:     *newTrans[ranges[i]].GetPath(),
				txIDs:   conflictTxs,
				ErrCode: ccurlcommon.ERR_ACCESS_CONFLICT,
			},
		)

		dict := common.MapFromArray(conflictTxs, true) //Conflict dict
		trans := common.CopyIf(newTrans[ranges[i]+offset:ranges[i+1]], func(v interfaces.Univalue) bool { return (*dict)[v.GetTx()] })

		if outOfLimits := (&Accumulator{}).CheckMinMax(trans); outOfLimits != nil {
			conflicts = append(conflicts, outOfLimits...)
		}
	}
	return conflicts
}