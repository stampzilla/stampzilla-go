package main

type StateValue struct {
	Type string
	Bool bool
	Int int
	String string
}

type State struct { /*{{{*/
	Values map[string]StateValue
} /*}}}*/

func (s *State) GetState() interface{} {
	return s
}
