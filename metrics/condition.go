package metrics

import "math"

var avgReplicasBySize []*condition
var avgReplicasByCreateTime []*condition
var fileCntByReplicaSize []*condition
var fileCntBySize []*condition
var fileCntBySizeNoneRep []*condition
var fileCntByCreateTime []*condition
var fileCntByExpireTime []*condition
var swokerRatio []*condition
var groupCntByMemberCnt []*condition
var groupCntByActiveCnt []*condition

func init() {
	avgReplicasBySize = []*condition{
		{"0~1KB", 0, 1024, 0},
		{"1KB~10KB", 1024, 10 * 1024, 0},
		{"10KB~100KB", 10 * 1024, 100 * 1024, 0},
		{"100KB~1MB", 100 * 1024, 1024 * 1024, 0},
		{"1MB~10MB", 1024 * 1024, 10 * 1024 * 1024, 0},
		{"10MB~100MB", 10 * 1024 * 1024, 100 * 1024 * 1024, 0},
		{"100MB~1GB", 100 * 1024 * 1024, 1024 * 1024 * 1024, 0},
		{">1GB", 1024 * 1024 * 1024, math.MaxUint64, 0},
	}

	avgReplicasByCreateTime = []*condition{
		{"<1小时", 0, 600, 0},
		{"1～3小时", 600, 1800, 0},
		{"3～12小时", 1800, 7200, 0},
		{"12小时～1天", 7200, 14400, 0},
		{"1天～7天", 14400, 7 * 14400, 0},
		{"7天～1个月", 7 * 14400, 30 * 14400, 0},
		{"1个月～6个月", 30 * 14400, 180 * 14400, 0},
		{">6个月", 180 * 14400, math.MaxUint64, 0},
	}

	fileCntByReplicaSize = []*condition{
		{"0(1)", 0, 0, 0},
		{"1～8(1.1)", 1, 8, 0},
		{"9～16(2)", 9, 16, 0},
		{"17～24(4)", 17, 24, 0},
		{"25～32(8)", 25, 32, 0},
		{"33～40(10)", 33, 40, 0},
		{"41～48(15)", 41, 48, 0},
		{"49～55(20)", 49, 55, 0},
		{"56～65(50)", 56, 65, 0},
		{"66～74(80)", 66, 74, 0},
		{"75～83(100)", 75, 83, 0},
		{"84～92(120)", 84, 92, 0},
		{"93～100(150)", 93, 100, 0},
		{"101～115(160)", 101, 115, 0},
		{"116～127(170)", 116, 127, 0},
		{"128～142(180)", 128, 142, 0},
		{"143～157(190)", 143, 157, 0},
		{"158～200(200)", 158, 200, 0},
		{">200(200)", 200, math.MaxUint64, 0},
	}

	fileCntBySize = []*condition{
		{"0~1KB", 0, 1024, 0},
		{"1KB~10KB", 1024, 10 * 1024, 0},
		{"10KB~100KB", 10 * 1024, 100 * 1024, 0},
		{"100KB~1MB", 100 * 1024, 1024 * 1024, 0},
		{"1MB~10MB", 1024 * 1024, 10 * 1024 * 1024, 0},
		{"10MB~30MB", 10 * 1024 * 1024, 30 * 1024 * 1024, 0},
		{"30MB~100MB", 30 * 1024 * 1024, 100 * 1024 * 1024, 0},
		{"100MB~300MB", 100 * 1024 * 1024, 300 * 1024 * 1024, 0},
		{"300MB~1GB", 300 * 1024 * 1024, 1024 * 1024 * 1024, 0},
		{">1GB", 1024 * 1024 * 1024, math.MaxUint64, 0},
	}

	fileCntBySizeNoneRep = []*condition{
		{"0~1KB", 0, 1024, 0},
		{"1KB~10KB", 1024, 10 * 1024, 0},
		{"10KB~100KB", 10 * 1024, 100 * 1024, 0},
		{"100KB~1MB", 100 * 1024, 1024 * 1024, 0},
		{"1MB~10MB", 1024 * 1024, 10 * 1024 * 1024, 0},
		{"10MB~30MB", 10 * 1024 * 1024, 30 * 1024 * 1024, 0},
		{"30MB~100MB", 30 * 1024 * 1024, 100 * 1024 * 1024, 0},
		{"100MB~300MB", 100 * 1024 * 1024, 300 * 1024 * 1024, 0},
		{"300MB~1GB", 300 * 1024 * 1024, 1024 * 1024 * 1024, 0},
		{">1GB", 1024 * 1024 * 1024, math.MaxUint64, 0},
	}

	fileCntByCreateTime = []*condition{
		{"<7天", 0, 100800, 0},
		{"7天~1个月", 100800, 432000, 0},
		{"1个月~3个月", 432000, 1296000, 0},
		{"3个月~6个月", 1296000, 2592000, 0},
		{"6个月~1年", 2592000, 5256000, 0},
		{"1年~2年", 5256000, 10512000, 0},
		{">2年", 10512000, math.MaxUint64, 0},
	}

	fileCntByExpireTime = []*condition{
		{"已过期", 0, 0, 0},
		{"<一个月", 0, 432000, 0},
		{"1个月~3个月", 432000, 1296000, 0},
		{"3个月~5个月", 1296000, 2160000, 0},
		{">5个月", 2160000, math.MaxUint64, 0},
	}

	swokerRatio = []*condition{
		{"0%", 0, 0, 0},
		{"0~0.1%", 0, 0.1, 0},
		{"0.1~0.3%", 0.1, 0.3, 0},
		{"0.3~1%", 0.3, 1, 0},
		{"1~3%", 1, 3, 0},
		{"3~10%", 3, 10, 0},
		{"10~30%", 10, 30, 0},
		{"30~50%", 30, 50, 0},
		{">50%", 50, math.MaxUint64, 0},
	}
	groupCntByMemberCnt = []*condition{
		{"0", 0, 0, 0},
		{"1", 1, 1, 0},
		{"2~5", 2, 5, 0},
		{"5~10", 5, 10, 0},
		{"10~20", 10, 20, 0},
		{"20~50", 20, 50, 0},
		{"50~100", 50, 100, 0},
		{">100", 100, math.MaxUint64, 0},
	}

	groupCntByActiveCnt = []*condition{
		{"0", 0, 0, 0},
		{"1", 1, 1, 0},
		{"2~5", 2, 5, 0},
		{"5~10", 5, 10, 0},
		{"10~20", 10, 20, 0},
		{"20~50", 20, 50, 0},
		{"50~100", 50, 100, 0},
		{">100", 100, math.MaxUint64, 0},
	}
}

type condition struct {
	name  string
	low   float64
	high  float64
	value float64
}
