package runner

type Runner interface {
	Start(nodes ...string) error
	Stop(nodes ...string) error
	Status() error
	Close()
}
