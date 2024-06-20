package metrics

import (
	log "github.com/ChainSafe/log15"
	"math"
	"statistic/chain"
	"statistic/db"
	"strconv"
)

type metricHandler func()

const PB uint64 = 1 << 50

const CommonInterval = 3600

var Handlers []struct {
	interval int
	handler  metricHandler
}
var slot uint64

func initHandler(interval int) {
	Handlers = []struct {
		interval int
		handler  metricHandler
	}{
		{interval, handlerAverageRepilicas},
		{interval, handlerFileAndSpower},
		{interval, handlerReplicaCntBySize},
		{interval, handlerReplicaCntByCreateTime},
		{interval, handlerFileCntByReplicas},
		{interval / 6, HandlerSlotFileCnt},
		{interval, handlerFileCntBySize},
		{interval, handlerFileCntByCreateTime},
		{interval, handlerFileCntByExpireTime},
	}
}

func initSlot(start uint64) {
	index, err := db.GetBlockNumber()
	if err != nil {
		index = 0
	}
	if index < start {
		index = start
	}
	slot = getSlot(index)
}

func getSlot(i uint64) uint64 {
	return i / chain.SlotSize * chain.SlotSize
}

//全网平均副本数
func handlerAverageRepilicas() {
	avg := db.AvgReplicas()
	chainMetric.AvgReplicas.Set(avg)
}

//全网文件数量、文件file_size和spower平均值
func handlerFileAndSpower() {
	count := db.FileCnt()
	chainMetric.FilesCnt.Set(float64(count))
	avgFileSize := db.AvgFileSize()
	avgSpower := db.AvgSpower()
	sumFileSize := avgFileSize * float64(count) / float64(PB)
	sumSpower := avgSpower * float64(count) / float64(PB)
	chainMetric.SumFileSpower.WithLabelValues("file_size").Set(sumFileSize)
	chainMetric.SumFileSpower.WithLabelValues("spower").Set(sumSpower)
	chainMetric.FileRatio.Set(avgSpower / avgFileSize)
	log.Info("handler File And Spower done")
}

//按文件大小统计平均副本数
func handlerReplicaCntBySize() {
	for _, c := range avgReplicasBySize {
		avg := db.AvgReplicasBySize(c.low, c.high)
		log.Debug("avg replicas by size", "label", c.name, "value", avg)
		c.value = avg
	}
	for _, c := range avgReplicasBySize {
		chainMetric.AvgReplicasBySize.WithLabelValues(c.name).Set(c.value)
	}
	log.Info("handlerReplicaCntBySize done")
}

//按创建时间统计平均副本数
func handlerReplicaCntByCreateTime() {
	now := chain.DefaultConn.GetLatestHeight()
	if now == 0 {
		return
	}
	for _, c := range avgReplicasByCreateTime {
		low := now - c.high
		if c.high == math.MaxUint64 {
			low = 0
		}
		high := now - c.low
		avg := db.AvgReplicasByCreateTime(low, high)
		log.Debug("avg replicas by create time", "label", c.name, "value", avg)
		c.value = avg
	}
	for _, c := range avgReplicasByCreateTime {
		chainMetric.AvgReplicasByCreateTime.WithLabelValues(c.name).Set(c.value)
	}
	log.Info("handlerReplicaCntByCreateTime done")
}

//按副本数量统计文件个数
func handlerFileCntByReplicas() {
	for _, c := range fileCntByReplicaSize {
		cnt := db.FileCntByReplicaSize(c.low, c.high)
		log.Debug("file count by replica size", "label", c.name, "value", cnt)
		c.value = float64(cnt)
	}
	for _, c := range fileCntByReplicaSize {
		chainMetric.FilesCntByReplicas.WithLabelValues(c.name).Set(c.value)
	}
	log.Info("handlerFileCntByReplicas done")
}

//HandlerSlotFileCnt 新增文件数
func HandlerSlotFileCnt() {
	bn, err := db.GetBlockNumber()
	if err != nil {
		return
	}
	if bn < slot {
		return
	}
	cnt, err := db.FileCntBySlot(slot)
	if err != nil {
		return
	}
	label := strconv.Itoa(int(slot))
	log.Debug("file count by replica size", "label", label, "value", cnt)
	chainMetric.FileCntBySlot.WithLabelValues(label).Set(float64(cnt))
	slot += chain.SlotSize
	log.Info("HandlerSlotFileCnt done")
}

//按文件大小统计文件个数
func handlerFileCntBySize() {
	for _, c := range fileCntBySize {
		cnt := db.FileCntBySize(c.low, c.high)
		log.Debug("file count by file size", "label", c.name, "value", cnt)
		c.value = float64(cnt)
	}
	for _, c := range fileCntBySize {
		chainMetric.FileCntBySize.WithLabelValues(c.name).Set(c.value)
	}
	log.Info("handlerFileCntBySize done")
}

//按创建时间统计文件个数
func handlerFileCntByCreateTime() {
	now := chain.DefaultConn.GetLatestHeight()
	if now == 0 {
		return
	}
	for _, c := range fileCntByCreateTime {
		low := now - c.high
		if c.high == math.MaxUint64 {
			low = 0
		}
		high := now - c.low
		cnt := db.FileCntByCreateTime(low, high)
		log.Debug("file count by create time", "label", c.name, "value", cnt)
		c.value = float64(cnt)
	}
	for _, c := range fileCntByCreateTime {
		chainMetric.FileCntByCreateTime.WithLabelValues(c.name).Set(c.value)
	}
	log.Info("handlerFileCntByCreateTime done")
}

//按文件过期时间统计文件个数
func handlerFileCntByExpireTime() {
	now := chain.DefaultConn.GetLatestHeight()
	if now == 0 {
		return
	}
	var low, high uint64
	for _, c := range fileCntByExpireTime {
		if c.high == c.low {
			low = 0
			high = now
		} else {
			low = c.low + now
			if c.high == math.MaxUint64 {
				high = math.MaxUint64
			} else {
				high = c.high + now
			}
		}
		cnt := db.FileCntByExpireTime(low, high)
		log.Debug("file count by expire time", "label", c.name, "value", cnt)
		c.value = float64(cnt)
	}
	for _, c := range fileCntByExpireTime {
		chainMetric.FileCntByExpireTime.WithLabelValues(c.name).Set(c.value)
	}
	log.Info("handlerFileCntByExpireTime done")
}
