package binary

import "github.com/davecgh/go-spew/spew"

type Installer struct {
}

func NewInstaller() *Installer {
	return &Installer{}
}

func (t *Installer) Prepare() error {
	return nil
}
func (t *Installer) Install(nodes ...string) {

	releases := getReleases()

	spew.Dump(releases)

}
func (t *Installer) Update(nodes ...string) {

}
