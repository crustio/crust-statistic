package chain

import (
	"fmt"
	log "github.com/ChainSafe/log15"
	"testing"
)

func ExampleStakeByIndex() {
	stop := make(chan int)
	conn := NewConnection(TestUrl, log.Root(), stop)
	err := conn.Connect()
	if err != nil {
		panic(err)
	}
	_, value, _ := GetStakeByIndex(conn)
	fmt.Printf("%v", value)
}

func ExampleTotalStake() {
	stop := make(chan int)
	conn := NewConnection(TestUrl, log.Root(), stop)
	err := conn.Connect()
	if err != nil {
		panic(err)
	}
	values, _ := GetTotalStakes(conn)
	fmt.Printf("%v", values)
}

func ExampleStakeLimit() {
	stop := make(chan int)
	conn := NewConnection(TestUrl, log.Root(), stop)
	err := conn.Connect()
	if err != nil {
		panic(err)
	}
	values, _ := GetTopStakeLimit(conn)
	fmt.Printf("%v", len(values))
}

func TestPayout(t *testing.T) {
	stop := make(chan int)
	conn := NewConnection(TestUrl, log.Root(), stop)
	err := conn.Connect()
	if err != nil {
		panic(err)
	}
	values, _ := GetStakingPayout(conn)
	fmt.Printf("%v \n", values)
	payouts, _ := GetAuthoringPayout(conn)
	fmt.Printf("%v \n", payouts)
	for _, value := range values {
		if v, ok := payouts[value.Index]; ok {
			value.Value += v
		}
	}
	for _, value := range values {
		println(value.Value)
	}
}

func TestPayoutByIndex(t *testing.T) {
	stop := make(chan int)
	conn := NewConnection(TestUrl, log.Root(), stop)
	err := conn.Connect()
	if err != nil {
		panic(err)
	}
	i, v, err := GetRewardByIndex(conn)
	if err != nil {
		panic(err)
	}
	println(i, v)
}
