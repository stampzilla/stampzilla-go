package main

import "github.com/sirupsen/logrus"

func main() {

	client := NewWebsocketClient()
	node := NewNode(client)

	err := node.Connect()

	if err != nil {
		logrus.Error(err)
		return
	}

	/*

		csr, err := node.GenerateCSR()
		if err != nil {
			logrus.Error(err)
			return
		}

		u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
		log.Printf("connecting to %s", u.String())

		err = node.ConnectWithRetry(u.String())
		if err != nil {
			logrus.Error(err)
		}

		msg, err := models.NewMessage("certificate-signing-request", string(csr))
		if err != nil {
			logrus.Error(err)
			return
		}

		node.Client.WriteJSON(msg)

		node.Wait()

		logrus.Info("node done...")
		logrus.Info("waiting for client to be done")
		node.Client.Wait()
		logrus.Info("client done")
	*/
}
