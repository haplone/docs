package main

import "fmt"

func main() {
	months := [...]string{1: "January", 2: "February", 3: "March",
		4: "April", 5: "May", 6: "June", 7: "July", 8: "August",
		9: "September", 10: "October", 11: "November", 12: "December"}

	q2 := months[4:7]
	summer := months[6:9]

	fmt.Println(q2)
	fmt.Println(summer)
	fmt.Println(&q2[2])
	fmt.Println(&summer[0])

	a := [...]int{1, 2, 3, 4, 5, 6}
	reverse(a[:])
	fmt.Println(a)

	fmt.Println(len(a))

	var runes []rune
	for _, r := range "Hello 世界" {
		runes = append(runes, r)
	}
	fmt.Printf("%q\n", runes)

	var as, ai, ay []int
	for i := 0; i < 10; i++ {
		ai = appendInt(as, i)
		ay = append(as, i)
		fmt.Printf("appendInt \t%d cap=%d\t%v\t%d\n", i, cap(ai), ai, &ai[0])
		fmt.Printf("append \t cap=%d\t%v\t%d\n", i, cap(ay), ay, &ay[0])
		as = ai
	}
}

func reverse(s []int) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func appendInt(x []int, y int) []int {
	var z []int
	zlen := len(x) + 1
	if zlen <= cap(x) {
		z = x[:zlen]
	} else {
		zcap := zlen
		if zcap < 2*len(x) {
			zcap = 2 * len(x)
		}
		z = make([]int, zlen, zcap)
		copy(z, x)
	}
	z[len(x)] = y
	return z
}
