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

type InstallSource uint8

const (
	Binaries = iota
	SourceCode
)

func New(s InstallSource) (Installer, error) {
	switch s {
	case Binaries:
		return binary.NewInstaller(), nil
	case SourceCode:
		return source.NewInstaller(), nil
	}
	return nil, fmt.Errorf("No installer with value %d is available", s)
}
