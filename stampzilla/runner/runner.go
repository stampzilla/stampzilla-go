package runner

type Runner interface {
	/* TODO: add methods */
	Start(nodes ...string) error
	Stop(nodes ...string) error
	Status() error
	Close()
}
