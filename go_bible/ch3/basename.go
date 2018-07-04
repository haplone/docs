package main

import "fmt"
import "strings"

func basename1(s string) string {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '/' {
			s = s[i+1:]
			break
		}
	}

	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			s = s[:i]
			break
		}
	}
	return s
}

func basename2(s string) string {
	slash := strings.LastIndex(s, "/")
	s = s[slash+1:]
	if dot := strings.LastIndex(s, "."); dot >= 0 {
		s = s[:dot]
	}
	return s
}

func main() {
	fmt.Println(basename1("a/d/g.go"))
	fmt.Println(basename1("d"))
	fmt.Println(basename1("23/345/sdf.java"))

	fmt.Println(basename2("a/d/g.go"))
	fmt.Println(basename2("d"))
	fmt.Println(basename2("23/345/sdf.java"))

}
