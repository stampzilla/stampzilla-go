package interfaces

type MelodyWriter interface {
	Write(msg []byte) error
}

type MelodySession interface {
	MelodyWriter
	Close() error
	CloseWithMsg(msg []byte) error
	Get(key string) (value interface{}, exists bool)
	IsClosed() bool
	// MustGet(key string) interface{} // removed in github.com/lesismal/melody fork
	Set(key string, value interface{})
	WriteBinary(msg []byte) error
}
