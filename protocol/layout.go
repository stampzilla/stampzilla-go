package protocol

type Layout struct { /*{{{*/
	Id      string
	Type    string
	Action  string
	Using   string
	Filter  []string
	Section string
} /*}}}*/

func NewLayout(id, atype, action, using string, filter []string, section string) *Layout {
	return &Layout{id, atype, action, using, filter, section}
}
