package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/RangelReale/osin"
	"github.com/gin-gonic/gin"
	"github.com/jonaz/gograce"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
)

func main() {
	config := &Config{}

	node := node.New("google-assistant")

	deviceList := devices.NewList()
	smartHomeHandler := NewSmartHomeHandler(node, deviceList)

	node.OnConfig(updatedConfig(config))
	wait := node.WaitForFirstConfig()

	err := node.Connect()
	if err != nil {
		logrus.Error(err)
		return
	}

	wait()
	node.On("devices", onDevices(config, deviceList))

	g := gin.Default()

	oauthConfig := osin.NewServerConfig()
	oauthConfig.AllowClientSecretInParams = true
	oauthConfig.AllowedAccessTypes = osin.AllowedAccessType{osin.AUTHORIZATION_CODE, osin.REFRESH_TOKEN}
	oauthStorage := NewJSONStorage()
	oauthStorage.LoadFromDisk("storage.json")
	oauthStorage.SetClient(config.ClientID, &osin.DefaultClient{
		Id:          config.ClientID,
		Secret:      config.ClientSecret,
		RedirectUri: "https://oauth-redirect.googleusercontent.com/r/" + config.ProjectID,
	})
	oauth2server := osin.NewServer(oauthConfig, oauthStorage)
	oauth2server.Logger = logrus.StandardLogger()

	g.GET("/authorize", authorize(oauth2server))
	g.POST("/authorize", authorize(oauth2server))

	g.GET("/token", token(oauth2server))
	g.POST("/token", token(oauth2server))

	g.POST("/", smartHomeHandler.smartHomeActionHandler(oauth2server))

	go func() {
		time.Sleep(5 * time.Second)
		logrus.Info("Syncing devices to google")
		requestSync(config.APIKey)
	}()

	srv, done := gograce.NewServerWithTimeout(1 * time.Second)

	srv.Handler = g
	srv.Addr = ":" + config.Port

	logrus.Error(srv.ListenAndServe())
	<-done
}

func onDevices(config *Config, deviceList *devices.List) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		list := devices.NewList()
		err := json.Unmarshal(data, list)
		if err != nil {
			return err
		}

		changes := 0
		for _, dev := range list.All() {
			old := deviceList.Get(dev.ID)
			if old == nil {
				deviceList.Add(dev)
				changes++
				continue
			}

			if old.Name != dev.Name {
				old.Lock()
				old.Name = dev.Name
				old.Unlock()
				changes++
			}

			if old.Alias != dev.Alias {
				old.Lock()
				old.Alias = dev.Alias
				old.Unlock()
				changes++
			}

			if old.Online != dev.Online {
				old.SetOnline(dev.Online)
				changes++
			}
		}

		toRemove := []devices.ID{}
		for _, v := range deviceList.All() {
			if dev := list.Get(v.ID); dev == nil {
				toRemove = append(toRemove, dev.ID)
			}
		}
		for _, id := range toRemove {
			changes++
			deviceList.Remove(id)
		}

		// Disabled because when stampzilla is restarted it will requestSync and google will delete all the devices in the cloud :(
		// TODO think of something smarter here! For now use the voice to ask google to sync devices.
		//if changes > 0 {
		//requestSync(config.APIKey)
		//}
		return nil
	}
}

func updatedConfig(config *Config) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		logrus.Debug("Received config from server:", string(data))
		return json.Unmarshal(data, &config)
	}
}

func requestSync(apiKey string) {
	if apiKey == "" {
		logrus.Info("we dont have apiKey. Skipping requestSync")
		return
	}
	u := fmt.Sprintf("https://homegraph.googleapis.com/v1/devices:requestSync?key=%s", apiKey)

	body := bytes.NewBufferString("{agentUserId: \"agentuserid\"}")
	req, err := http.NewRequest("POST", u, body)
	if err != nil {
		logrus.Error("requestsync: ", err)
		return
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.Error("requestsync: ", err)
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Error("requestsync: ", err)
		return
	}

	logrus.Debug("requestSync response:", string(data))
}
