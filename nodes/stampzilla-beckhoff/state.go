package main

import (
	"github.com/stamp/goADS"
)


type StateValue struct {
	Type   string
	Bool   bool
	Int    int
	String string

	symbol *goADS.ADSSymbol
}

type State struct { /*{{{*/
	Values map[string]StateValue
} /*}}}*/
