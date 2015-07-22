package main

type StateValue struct {
	Type   string
	Bool   bool
	Int    int
	String string
}

type State struct { /*{{{*/
	Values map[string]StateValue
} /*}}}*/
