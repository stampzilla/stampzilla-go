package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/RangelReale/osin"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"
	"github.com/stampzilla/stampzilla-go/protocol/devices"
)

// VERSION is the app version
var VERSION = "dev"

// BUILDDATE is then the binary was built
var BUILDDATE = ""
var listenPort string
var debug bool
var standalone bool
var nodespecific = newNodeSpecificConfig()

func init() {
	flag.StringVar(&listenPort, "listenport", "80", "Port to listen to. Must be 80 for Google Home to work")
	flag.BoolVar(&debug, "debug", false, "Debug. Without this we dont print other than errors. Optimized not to wear on raspberry pi SD card.")
	flag.BoolVar(&standalone, "standalone", false, "Run standalone without communicating with stampzilla-server Host:Port configured in config.json.")
}

func main() {
	flag.Parse()

	if debug {
		log.Println("Setting logrus level to DEBUG")
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		log.Println("Setting logrus level to INFO")
		logrus.SetLevel(logrus.InfoLevel)
	}

	config := basenode.NewConfig()

	basenode.SetConfig(config)

	node := protocol.NewNode("google-assistant")
	node.Version = VERSION
	node.BuildDate = BUILDDATE

	//devices := NewDevices()
	err := config.NodeSpecific(&nodespecific)
	if err != nil {
		log.Println(err)
	}

	if nodespecific.ListenPort != "" && listenPort == "80" {
		listenPort = nodespecific.ListenPort
	}

	go func() {
		for range time.NewTicker(60 * time.Second).C {
			if debug {
				log.Println("Syncing devices from server")
			}
			syncDevicesFromServer(config, nodespecific)

		}
	}()
	connection := basenode.Connect()
	go monitorState(node, connection)

	g := gin.Default()

	oauthConfig := osin.NewServerConfig()
	oauthConfig.AllowClientSecretInParams = true
	oauthConfig.AllowedAccessTypes = osin.AllowedAccessType{osin.AUTHORIZATION_CODE, osin.REFRESH_TOKEN}
	oauthStorage := NewTestStorage()
	oauthStorage.LoadFromDisk("storage.json")
	oauthStorage.SetClient(nodespecific.ClientID, &osin.DefaultClient{
		Id:          nodespecific.ClientID,
		Secret:      nodespecific.ClientSecret,
		RedirectUri: "https://oauth-redirect.googleusercontent.com/r/" + nodespecific.ProjectID,
	})
	oauth2server := osin.NewServer(oauthConfig, oauthStorage)
	oauth2server.Logger = logrus.New()

	g.GET("/authorize", authorize(oauth2server))
	g.POST("/authorize", authorize(oauth2server))

	g.GET("/token", token(oauth2server))
	g.POST("/token", token(oauth2server))

	g.POST("/", smartHomeActionHandler(oauth2server))

	go func() {
		time.Sleep(5 * time.Second)
		log.Println("Syncing devices from server")
		syncDevicesFromServer(config, nodespecific)
	}()

	log.Println(g.Run(":" + listenPort))
}

func authorize(oauth2server *osin.Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		resp := oauth2server.NewResponse()
		defer resp.Close()

		if ar := oauth2server.HandleAuthorizeRequest(resp, c.Request); ar != nil {

			// HANDLE LOGIN PAGE HERE
			if !handleLoginPage(ar, c.Writer, c.Request) {
				return
			}
			ar.UserData = struct{ Login string }{Login: "test"}
			ar.Authorized = true
			oauth2server.FinishAuthorizeRequest(resp, c.Request, ar)
		}
		if resp.IsError && resp.InternalError != nil {
			logrus.Errorf("ERROR: %#v\n", resp.InternalError)
		}
		osin.OutputJSON(resp, c.Writer, c.Request)
	}
}
func token(oauth2server *osin.Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		resp := oauth2server.NewResponse()
		defer resp.Close()

		if ar := oauth2server.HandleAccessRequest(resp, c.Request); ar != nil {
			ar.Authorized = true
			oauth2server.FinishAccessRequest(resp, c.Request, ar)
		}
		if resp.IsError && resp.InternalError != nil {
			logrus.Errorf("ERROR: %#v\n", resp.InternalError)
		}
		osin.OutputJSON(resp, c.Writer, c.Request)
	}
}

func syncDevicesFromServer(config *basenode.Config, ns *nodeSpecificConfig) bool {
	didChange := false

	serverDevs, err := fetchDevices(config, ns)
	if err != nil {
		log.Println(err)
		return false
	}

outer:
	for uuid, sdev := range serverDevs.All() {
		for _, v := range ns.Devices() {
			if v.ID == uuid {
				if debug {
					log.Printf("Already have device: %s. Do not add again.\n", sdev.Name)
				}
				if v.Name != sdev.Name {
					v.Name = sdev.Name
					didChange = true
				}
				continue outer
			}
		}

		//Skip non controllable devices
		if sdev.Type != "lamp" && sdev.Type != "dimmableLamp" {
			continue
		}

		//We dont have the device so we add it
		baseURL := fmt.Sprintf("http://%s:%s/api/nodes/", config.Host, ns.Port)
		dev := &Device{
			Name: sdev.Name,
			Url: &URL{
				Level: baseURL + sdev.Node + "/cmd/level/" + sdev.Id + "/%d",
				On:    baseURL + sdev.Node + "/cmd/on/" + sdev.Id,
				Off:   baseURL + sdev.Node + "/cmd/off/" + sdev.Id,
			},
			ID: uuid,
		}

		didChange = true
		ns.AddDevice(dev)
	}

	//Dont save file if no new devices are found
	if !didChange {
		return false
	}

	data, err := json.Marshal(ns)
	if err != nil {
		log.Println(err)
		return false
	}
	raw := json.RawMessage(data)
	config.Node = &raw
	basenode.SaveConfigToFile(config)
	requestSync()
	//TODO request sync to google if new devices found
	return true
}

func requestSync() {
	if nodespecific.APIKey == "" {
		return
	}
	u := fmt.Sprintf("https://homegraph.googleapis.com/v1/devices:requestSync?key=%s", nodespecific.APIKey)

	body := bytes.NewBufferString("{agent_user_id: \"agentuserid\"}")
	req, err := http.NewRequest("POST", u, body)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	logrus.Info("requestSync response:", string(data))
}

func fetchDevices(config *basenode.Config, ns *nodeSpecificConfig) (*devices.Map, error) {
	//TODO use nodespecific config
	url := fmt.Sprintf("http://%s:%s/api/devices", config.Host, ns.Port)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	devmap := devices.NewMap()
	err = json.NewDecoder(resp.Body).Decode(&devmap)
	if err != nil {
		return nil, err
	}
	return devmap, nil

}

// WORKER that monitors the current connection state
func monitorState(node *protocol.Node, connection basenode.Connection) {
	for s := range connection.State() {
		switch s {
		case basenode.ConnectionStateConnected:
			connection.Send(node.Node())
		case basenode.ConnectionStateDisconnected:
		}
	}
}

func handleLoginPage(ar *osin.AuthorizeRequest, w http.ResponseWriter, r *http.Request) bool {
	r.ParseForm()
	if r.Method == "POST" && r.Form.Get("login") == "test" && r.Form.Get("password") == "test" {
		return true
	}

	w.Write([]byte("<html><body>"))

	w.Write([]byte(fmt.Sprintf("LOGIN %s (use test/test)<br/>", ar.Client.GetId())))
	w.Write([]byte(fmt.Sprintf("<form action=\"/authorize?%s\" method=\"POST\">", r.URL.RawQuery)))

	w.Write([]byte("Login: <input type=\"text\" name=\"login\" /><br/>"))
	w.Write([]byte("Password: <input type=\"password\" name=\"password\" /><br/>"))
	w.Write([]byte("<input type=\"submit\"/>"))

	w.Write([]byte("</form>"))

	w.Write([]byte("</body></html>"))

	return false
}
