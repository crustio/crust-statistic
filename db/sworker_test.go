package db

import "testing"

func TestListAnchor(t *testing.T) {
	InitMysql(getConfig())
	res, err := GetVersionCnt()
	if err != nil {
		panic(err)
	}
	for _, r := range res {
		println(r.Code, r.Cnt)
	}
}

func ExampleTopGroups() {
	InitMysql(getConfig())
	res, err := GetTopGroups()
	if err != nil {
		panic(err)
	}
	for _, r := range res {
		println(r.Spower)
	}
}
