package test

import (
	"pjt/internal/utils"
	"testing"
)

type TestStruct struct {
	First string
	Last  string
	Mp    map[int]*TestStruct1
}

type TestStruct1 struct {
	Id   int
	Val1 float32
	Val2 string
	Val3 byte
}

func TestUtils(t *testing.T) {
	p1 := TestStruct{
		First: "James",
		Last:  "Bond",
		Mp:    make(map[int]*TestStruct1),
	}
	p1.Mp[1] = &TestStruct1{
		Id:   2,
		Val1: 3.2,
		Val2: "sdf",
		Val3: 0x02,
	}

	p1.Mp[2] = &TestStruct1{
		Id:   3,
		Val1: 4.2,
		Val2: "ddsdf",
		Val3: 0x03,
	}

	p2 := TestStruct{}

	utils.DeepCopy(&p1, &p2)

	t.Log(p2.First, p2.Last)
	for k, v := range p2.Mp {
		t.Logf("key: %+v, val: %+v", k, v)
	}
}
