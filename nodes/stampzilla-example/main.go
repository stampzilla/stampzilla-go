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

	node.Wait()
	node.Client.Wait()
}
