package main

import (
	"bufio"
	"io"
	"log"
)

func cat(r io.Reader, w io.Writer) {
	var (
		br *bufio.Reader
		e  error
	)

	br = bufio.NewReader(r)

	for {
		_, e = br.WriteTo(w)
		if e != nil {
			log.Println(e)
			return
		}
	}
}
