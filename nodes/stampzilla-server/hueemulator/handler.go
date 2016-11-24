package hueemulator

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

var handlerMap map[string]huestate

func init() {
	handlerMap = make(map[string]huestate)
	upnpTemplateInit()
}

func SetLogger(w io.Writer) {
	log.SetOutput(w)
}

func ListenAndServe(addr string) error {
	//router := httprouter.New()
	router := gin.Default()

	router.GET(upnp_uri, upnpSetup(addr))

	router.GET("/api/:userId/lights", getLightsList)
	router.PUT("/api/:userId/lights/:lightId/state", setLightState)
	router.GET("/api/:userId/lights/:lightId", getLightInfo)
	router.POST("/api", func(c *gin.Context) {
		defer c.Request.Body.Close()
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Println(err)
		}
		log.Println(string(body))
		c.Writer.WriteString(`[{"success":{"username": "83b7780291a6ceffbe0bd049104df"}}]`)
	})

	go upnpResponder(addr, upnp_uri)
	return http.ListenAndServe(addr, requestLogger(router))
}

// Handler:
// 	state is the state of the "light" after the handler function
//  if error is set to true echo will reply with "sorry the device is not responding"
type Handler func(Request, *Response)

func Handle(deviceName string, h Handler) {
	log.Println("[HANDLE]", deviceName)
	handlerMap[deviceName] = huestate{
		Handler: h,
		OnState: false,
	}
}

func requestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("[WEB]", r.RemoteAddr, r.Method, r.URL)
		//		log.Printf("\t%+v\n", r)
		h.ServeHTTP(w, r)
	})
}
