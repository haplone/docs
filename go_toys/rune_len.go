package main

import "log"

func main() {
	// 使用rune处理中文
	str := "你好sdf水电费"

	b := []rune(str)

	log.Println(len(str))
	log.Println(len(b))
	log.Println(string(b[:len(b)-2]))

	if len(b) >=8{
		log.Println(string(b[:8]))
	}
}
