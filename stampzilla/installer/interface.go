package installer

import (
	"fmt"

	"github.com/stampzilla/stampzilla-go/stampzilla/installer/binary"
	"github.com/stampzilla/stampzilla-go/stampzilla/installer/source"
)

type Installer interface {
	Prepare() error
	Install(...string)
	Update(...string)
}

func New(t string) (Installer, error) {
	switch t {
	case "binaries":
		return newFromBinaries(), nil
	case "source":
		return newFromSource(), nil
	}

	return nil, fmt.Errorf("No installer named \"%s\" is available", t)
}

func newFromBinaries() Installer {
	return binary.NewInstaller()
}
func newFromSource() Installer {
	return source.NewInstaller()
}
