package hueemulator

import (
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
)

var handlerMap map[int]*huestate
var debug bool
var handlerMapLock sync.Mutex

func init() {
	log.SetOutput(ioutil.Discard)
	handlerMap = make(map[int]*huestate)
	upnpTemplateInit()
}

func SetLogger(w io.Writer) {
	log.SetOutput(w)
}

func SetDebug(d bool) {
	debug = d
}

func ListenAndServe(addr string) error {
	log.Println("Listening to: ", addr)
	//router := httprouter.New()
	//router := gin.Default()
	router := gin.New()

	if debug {
		router.Use(gin.Logger())
	}
	router.Use(gin.Recovery())

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
type Handler func(Request) error

func Handle(id int, deviceName string, h Handler) {
	log.Println("[HANDLE]", deviceName)
	handlerMapLock.Lock()
	handlerMap[id] = &huestate{
		Handler: h,
		Light:   initLight(id, deviceName),
		Id:      id,
	}
	handlerMapLock.Unlock()
}

func requestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if debug {
			log.Println("[WEB]", r.RemoteAddr, r.Method, r.URL)
		}
		h.ServeHTTP(w, r)
	})
}
func GetPrimaryIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
