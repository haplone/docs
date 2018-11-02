package tree

import (
	"testing"
	"fmt"
	//"log"
)

func TestAdd(t *testing.T) {
	n := getTree()
	tl := getTarget()
	//fmt.Println(n.t())

	if rl := n.t(); !checkSlice(rl, tl) {
		t.Error("add node error ,should be ", tl, " but got ", rl)
	}
}

func testAppend(t *testing.T) {
	tl := getTarget()

	for idx, t := range tl {
		fmt.Printf("\t idx:%d\t: %d \n", idx, t)
		var t1, t2 []int = make([]int, idx), make([]int, len(tl)-idx-1)
		copy(t1, tl[:idx])
		copy(t2, tl[idx+1:])

		fmt.Println("tl:\t", tl)
		fmt.Println("tls:\t", t1)
		fmt.Println("tle:\t", t2)

		tt := append(t1, t2...)
		fmt.Println("tlr:\t", tt)
	}

}

func TestRemove(t *testing.T) {
	tl := getTarget()
	//tl := []int{1, 2, 3, 4, 6, 30, 34, 56, 87, 89, 234, 345}

	for idx, c := range tl {
		if c != 30 {
			t1, t2 := make([]int, idx), make([]int, len(tl)-idx-1)
			copy(t1, tl[:idx])
			copy(t2, tl[idx+1:])
			tlt := append(t1, t2...)

			n := getTree()
			n.remove(c)
			//t.Log("rm ", c)
			//t.Log(tl)
			//t.Log(n.t())
			if rl := n.t(); !checkSlice(rl, tlt) {
				t.Error("remove node error ,should be ", tlt, " but got ", rl)
			}
		}
	}

}

func TestSearch(t *testing.T) {
	tl := getTarget()

	for _, c := range tl {
		n := getTree()

		if cn := n.Search(c); cn.data != c {
			t.Error(" get node error, should get ", c, " but got ", cn.data)
		}

	}
}

func getTarget() []int {
	tl := []int{1, 2, 3, 4, 6, 30, 34, 56, 87, 89, 234, 345}
	return tl
}
func getTree() *BstNode {
	list := []int{1, 345, 4, 34, 3, 6, 2, 6, 3, 56, 234, 6, 87, 89}
	var n = BstNode{
		data: 30,
	}
	for _, i := range list {
		//fmt.Println("insert ", i)
		n.Add(i)
	}

	return &n
}

func checkSlice(s, t []int) bool {
	if len(s) != len(t) {
		return false
	}
	for i, l := 0, len(s); i < l; i++ {
		if s[i] != t[i] {
			return false
		}
	}
	return true
}
