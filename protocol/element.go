package protocol

const (
	ElementTypeText        = 0
	ElementTypeButton      = 1
	ElementTypeToggle      = 2
	ElementTypeSlider      = 3
	ElementTypeColorPicker = 4
)

type Element struct {
	Type     int
	Name     string
	Command  *Command
	Feedback string
}
