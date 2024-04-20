package main

import(
	"fmt"
)

type graf struct {
	nilai string
	list<string> *tetangga
}

func (g *graf) addTetangga