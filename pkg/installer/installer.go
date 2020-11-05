package installer

import (
	"fmt"

	"github.com/stampzilla/stampzilla-go/v2/pkg/installer/binary"
	"github.com/stampzilla/stampzilla-go/v2/pkg/installer/source"
)

type Installer interface {
	Prepare() error
	Install(...string) error
	Update(...string) error
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
