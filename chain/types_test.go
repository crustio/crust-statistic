package chain

import (
	"gotest.tools/assert"
	"testing"
)

const TestHex = "0x0a94123ede29475590001f364d58ce88a83647d63bfd7b38e65e0ae58a78f804"

const TestAcc = "cTGaWK3pHpExP1cyrZNArNfWAWidtPCZx29mTNbDY3Yy9R9HG"

func TestConvert(t *testing.T) {
	val := convertAccount(TestHex)
	assert.Equal(t, val, TestAcc)
}
