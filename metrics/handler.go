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

var sworkerActive = false

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
		{interval, handlerSwoker},
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
	avg, err := db.AvgReplicas()
	if err != nil {
		log.Error("get avg replicas error", "err", err)
		return
	}
	chainMetric.avgReplicas.Set(avg)
}

//全网文件数量、文件file_size和spower平均值
func handlerFileAndSpower() {
	count, err := db.FileCnt()
	if err != nil {
		log.Error("get file count error", "err", err)
		return
	}
	chainMetric.filesCnt.Set(float64(count))
	avgFileSize, err := db.AvgFileSize()
	if err != nil {
		log.Error("get avg file size error", "err", err)
		return
	}
	avgSpower, err := db.AvgSpower()
	if err != nil {
		log.Error("get avg spower size error", "err", err)
		return
	}
	sumFileSize := avgFileSize * float64(count) / float64(PB)
	sumSpower := avgSpower * float64(count) / float64(PB)
	chainMetric.sumFileSpower.WithLabelValues("file_size").Set(sumFileSize)
	chainMetric.sumFileSpower.WithLabelValues("spower").Set(sumSpower)
	chainMetric.fileRatio.Set(avgSpower / avgFileSize)
	log.Info("handler File And Spower done")
}

//按文件大小统计平均副本数
func handlerReplicaCntBySize() {
	for _, c := range avgReplicasBySize {
		avg, err := db.AvgReplicasBySize(uint64(c.low), uint64(c.high))
		if err != nil {
			log.Error("get avg replicas by size error", "err", err)
			return
		}
		log.Debug("avg replicas by size", "label", c.name, "value", avg)
		c.value = avg
	}
	for _, c := range avgReplicasBySize {
		chainMetric.avgReplicasBySize.WithLabelValues(c.name).Set(c.value)
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
		low := now - uint64(c.high)
		if c.high == math.MaxUint64 {
			low = 0
		}
		high := now - uint64(c.low)
		avg, err := db.AvgReplicasByCreateTime(low, high)
		if err != nil {
			log.Error("get avg replicas by create time error", "err", err)
			return
		}
		log.Debug("avg replicas by create time", "label", c.name, "value", avg)
		c.value = avg
	}
	for _, c := range avgReplicasByCreateTime {
		chainMetric.avgReplicasByCreateTime.WithLabelValues(c.name).Set(c.value)
	}
	log.Info("handlerReplicaCntByCreateTime done")
}

//按副本数量统计文件个数
func handlerFileCntByReplicas() {
	for _, c := range fileCntByReplicaSize {
		cnt, err := db.FileCntByReplicaSize(uint64(c.low), uint64(c.high))
		if err != nil {
			log.Error("get file count by replicas size error", "err", err)
			return
		}
		log.Debug("file count by replica size", "label", c.name, "value", cnt)
		c.value = float64(cnt)
	}
	for _, c := range fileCntByReplicaSize {
		chainMetric.filesCntByReplicas.WithLabelValues(c.name).Set(c.value)
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
	chainMetric.fileCntBySlot.WithLabelValues(label).Set(float64(cnt))
	slot += chain.SlotSize
	log.Info("HandlerSlotFileCnt done")
}

//按文件大小统计文件个数
func handlerFileCntBySize() {
	for _, c := range fileCntBySize {
		cnt, err := db.FileCntBySize(uint64(c.low), uint64(c.high))
		if err != nil {
			log.Error("get file count by size error", "err", err)
			return
		}
		log.Debug("file count by file size", "label", c.name, "value", cnt)
		c.value = float64(cnt)
	}
	for _, c := range fileCntBySize {
		chainMetric.fileCntBySize.WithLabelValues(c.name).Set(c.value)
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
		low := now - uint64(c.high)
		if c.high == math.MaxUint64 {
			low = 0
		}
		high := now - uint64(c.low)
		cnt, err := db.FileCntByCreateTime(low, high)
		if err != nil {
			log.Error("get file count by create time error", "err", err)
			return
		}
		log.Debug("file count by create time", "label", c.name, "value", cnt)
		c.value = float64(cnt)
	}
	for _, c := range fileCntByCreateTime {
		chainMetric.fileCntByCreateTime.WithLabelValues(c.name).Set(c.value)
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
			low = uint64(c.low) + now
			if c.high == math.MaxUint64 {
				high = math.MaxUint64
			} else {
				high = uint64(c.high) + now
			}
		}
		cnt, err := db.FileCntByExpireTime(low, high)
		if err != nil {
			log.Error("get file count by expire time error", "err", err)
			return
		}
		log.Debug("file count by expire time", "label", c.name, "value", cnt)
		c.value = float64(cnt)
	}
	for _, c := range fileCntByExpireTime {
		chainMetric.fileCntByExpireTime.WithLabelValues(c.name).Set(c.value)
	}
	log.Info("handlerFileCntByExpireTime done")
}

func handlerSwoker() {
	if sworkerActive {
		return
	}
	sworkerActive = true
	err := db.ClearSworker()
	if err != nil {
		log.Error("clear tmp data  error", "err", err)
		sworkerActive = false
		return
	}
	all, active, err := chain.GetAllSworkReports(chain.DefaultConn)
	if err != nil {
		log.Error("get swork report error", "err", err)
		sworkerActive = false
		return
	}
	log.Info("get swork report done")
	go handlerStorage(all, active)
	go handlerSwokerByRatio()
	go handlerSworkerVersion()
	err = chain.GetGroupInfo(chain.DefaultConn)
	if err != nil {
		log.Error("get group info error", "err", err)
		return
	}
	go handlerGroupCnt()
	go handlerGroupByMemberCnt()
	go handlerGroupByActiveCnt()
	sworkerActive = false
}

func handlerStorage(all, active int) {
	free, err := db.SumFree()
	if err != nil {
		log.Error("get storage free error", "err", err)
		return
	}
	fileSize, err := db.SumFileSize()
	if err != nil {
		log.Error("get storage file size error", "err", err)
		return
	}
	chainMetric.sworkerCnt.WithLabelValues("all").Set(float64(all))
	chainMetric.sworkerCnt.WithLabelValues("active").Set(float64(active))
	freePB := free / float64(PB)
	fileSizePB := fileSize / float64(PB)
	allPB := freePB + fileSizePB
	chainMetric.storageSize.WithLabelValues("all").Set(allPB)
	chainMetric.storageSize.WithLabelValues("free").Set(freePB)
	chainMetric.storageSize.WithLabelValues("used").Set(fileSizePB)
}

func handlerSwokerByRatio() {
	for _, c := range swokerRatio {
		cnt, err := db.NodeCntByRatio(c.low, c.high)
		if err != nil {
			log.Error("get swoker count by ratio error", "err", err)
			return
		}
		log.Debug("sworker count by ratio", "label", c.name, "value", cnt)
		c.value = float64(cnt)
	}
	for _, c := range swokerRatio {
		chainMetric.sworkerCntByRatio.WithLabelValues(c.name).Set(c.value)
	}
	log.Info("SworkerCntByRatio done")
}

func handlerGroupCnt() {
	all, err := db.GroupCnt()
	if err != nil {
		log.Error("get group cnt error", "err", err)
		return
	}
	active, err := db.GroupActiveCnt()
	if err != nil {
		log.Error("get group active cnt error", "err", err)
		return
	}
	chainMetric.groupCnt.WithLabelValues("all").Set(float64(all))
	chainMetric.groupCnt.WithLabelValues("active").Set(float64(active))

	avgMember, err := db.AvgMembers()
	if err != nil {
		log.Error("get avg member cnt error", "err", err)
		return
	}
	avgActiveMember, err := db.AvgActiveMembers()
	if err != nil {
		log.Error("get avg active member cnt error", "err", err)
		return
	}
	chainMetric.avgSworkerCntByGroup.WithLabelValues("all").Set(avgMember)
	chainMetric.avgSworkerCntByGroup.WithLabelValues("active").Set(avgActiveMember)
}

func handlerGroupByMemberCnt() {
	for _, c := range groupCntByMemberCnt {
		cnt, err := db.GroupCntByAll(uint64(c.low), uint64(c.high))
		if err != nil {
			log.Error("get group cnt by member cnt error", "err", err)
			return
		}
		log.Debug("group count by member cnt", "label", c.name, "value", cnt)
		c.value = float64(cnt)
	}
	for _, c := range groupCntByMemberCnt {
		chainMetric.groupCntBySworkerCnt.WithLabelValues(c.name).Set(c.value)
	}
	log.Info("GroupCntBySworkerCnt done")
}

func handlerGroupByActiveCnt() {
	for _, c := range groupCntByActiveCnt {
		cnt, err := db.GroupCntByActive(uint64(c.low), uint64(c.high))
		if err != nil {
			log.Error("get group cnt by active member cnt error", "err", err)
			return
		}
		log.Debug("group count by active cnt", "label", c.name, "value", cnt)
		c.value = float64(cnt)
	}
	for _, c := range groupCntByActiveCnt {
		chainMetric.groupCntByActiveSworkerCnt.WithLabelValues(c.name).Set(c.value)
	}
	log.Info("GroupCntByActiveSworkerCnt done")
}

func handlerSworkerVersion() {
	err := chain.GetPubKeys(chain.DefaultConn)
	if err != nil {
		log.Error("get pub keys error", "err", err)
		return
	}
	codes, err := db.GetVersionCnt()
	if err != nil {
		log.Error("db version cnt error", "err", err)
		return
	}

	versionCnt := make(map[string]int)
	for _, code := range codes {
		if version, ok := versionMap[code.Code]; ok {
			versionCnt[version] += code.Cnt
		} else {
			versionCnt[versionMap["0x"]] += code.Cnt
		}
	}

	for version, cnt := range versionCnt {
		chainMetric.sworkerByVersion.WithLabelValues(version).Set(float64(cnt))
	}
	log.Info("sworker version done")
}
