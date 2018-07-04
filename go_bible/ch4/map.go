package main

import "fmt"
import "sort"

func main() {
	ages := make(map[string]int)

	ages2 := map[string]int{
		"alice":   31,
		"charlie": 34,
	}

	fmt.Println(ages)
	fmt.Println(ages2)

	for name, age := range ages2 {
		fmt.Printf(" people name: %s \t,age:%d \n", name, age)
	}

	var names []string
	for name := range ages2 {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Printf(" people after sort name: %s\tage: %d\n", name, ages2[name])
	}

	delete(ages2, "a")
	delete(ages2, "charlie")
	fmt.Println(ages2)
	delete(ages2, "alice")
	fmt.Println(ages2["alice"])

}
