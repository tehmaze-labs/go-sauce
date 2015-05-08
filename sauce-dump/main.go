package main

import (
	"flag"
	"fmt"
)

func main() {
	flag.Parse()

	for _, filename := range flag.Args() {
		r, err := suace.Parse(filename)
		if err != nil {
			fmt.Printf("%s: error %v\n", filename, err)
			continue
		}
		if r == nil {
			fmt.Printf("%s: no SAUCE record\n", filename)
			continue
		}

		fmt.Println(filename)
		r.Dump()
	}
}
