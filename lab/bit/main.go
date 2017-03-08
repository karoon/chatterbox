package main

import "fmt"

func main() {
	pp(12)
	pp(10)
	pp(12 & 10)
	pp(16)
	pp(0xF0)
	pp(16 & 0xF0)
	pp(0x08)
	pp(16 & 0x08)
}

func pp(n interface{}) {
	fmt.Println(n)
	fmt.Printf("decimal: %s\n", fmt.Sprintf("%d", n))
	fmt.Printf("hex: %x\n", n)
	fmt.Printf("binary: %b\n", n)
	// xb := 1 > 0
	// log.Printf(xb)
	fmt.Println("--------------")
}
