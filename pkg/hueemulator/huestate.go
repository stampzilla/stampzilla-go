package hueemulator

type huestate struct {
	Handler Handler
	//OnState bool
	Light *light
	Id    int
}
