package main

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	log "github.com/cihub/seelog"
	"github.com/gin-gonic/gin"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/logic"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/servernode"
	"github.com/stampzilla/stampzilla-go/protocol"
)

type WebHandler struct {
	Logic          *logic.Logic          `inject:""`
	Scheduler      *logic.Scheduler      `inject:""`
	Nodes          *serverprotocol.Nodes `inject:""`
	NodeServer     *NodeServer           `inject:""`
	ActionsService *logic.ActionService        `inject:""`
}

func (wh *WebHandler) GetNodes(c *gin.Context) {
	c.JSON(200, wh.Nodes.All())
}

func (wh *WebHandler) GetNode(c *gin.Context) {
	if n := wh.Nodes.Search(c.Param("id")); n != nil {
		c.JSON(200, n)
		return
	}
	c.String(404, "{}")
}

func (wh *WebHandler) CommandToNodePut(c *gin.Context) {
	id := c.Param("id")
	log.Info("Sending command to:", id)
	requestJsonPut, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Error(err)
		c.String(500, "Error")
	}
	log.Info("Command:", string(requestJsonPut))

	node := wh.Nodes.Search(id)
	if node == nil {
		log.Debug("NODE: ", node)
		c.String(404, "Node not found")
	}

	node.Write(requestJsonPut)
	c.JSON(200, protocol.Command{Cmd: "testresponse"})
}

func (wh *WebHandler) CommandToNodeGet(c *gin.Context) {
	id := c.Param("id")
	node := wh.Nodes.Search(id)
	if node == nil {
		log.Debug("NODE: ", node)
		c.String(404, "Node not found")
		return
	}

	log.Info("Sending command to:", id)

	// Split on / to add arguments
	p := strings.Split(c.Param("cmd"), "/")
	cmd := protocol.Command{Cmd: p[1], Args: p[2:]}

	jsonCmd, err := json.Marshal(cmd)
	if err != nil {
		log.Error(err)
		c.String(500, "Failed json marshal")
		return
	}
	log.Info("Command:", string(jsonCmd))
	node.Write(jsonCmd)
	c.JSON(200, protocol.Command{Cmd: "testresponse"})
}

func (wh *WebHandler) GetActions(c *gin.Context) {
	c.JSON(200, wh.ActionsService.Get())
}
func (wh *WebHandler) ReloadActions(c *gin.Context) {
	wh.ActionsService.Start()
	c.JSON(200, wh.ActionsService.Get())
}
func (wh *WebHandler) GetRules(c *gin.Context) {
	c.JSON(200, wh.Logic.Rules())
}
func (wh *WebHandler) GetRunRules(c *gin.Context) {
	id := c.Param("id")
	action := c.Param("action")
	for _, rule := range wh.Logic.Rules() {
		if rule.Uuid() == id {
			switch action {
			case "enter":
				rule.RunEnter()
				c.JSON(200, "ok")
				return
			case "exit":
				rule.RunExit()
				c.JSON(200, "ok")
				return

			}
		}
	}
	c.String(404, "Rule not found")
}

func (wh *WebHandler) GetScheduleTasks(c *gin.Context) {
	c.JSON(200, wh.Scheduler.Tasks())
}
func (wh *WebHandler) GetScheduleEntries(c *gin.Context) {
	c.JSON(200, wh.Scheduler.Cron.Entries())
}

func (wh *WebHandler) GetScheduleReload(c *gin.Context) {
	wh.Scheduler.Reload()
	c.JSON(200, wh.Scheduler.Tasks())
}
func (wh *WebHandler) GetReload(c *gin.Context) {
	wh.Logic.RestoreRulesFromFile("rules.json")
	c.JSON(200, wh.Logic.Rules())
}

func (wh *WebHandler) GetServerTrigger(c *gin.Context) {

	if node, ok := wh.Nodes.ByName("server").(*servernode.Node); ok {
		node.Set(c.Param("key"), c.Param("value"))
		wh.NodeServer.updateState(node.LogicChannel(), node)
		node.Reset(c.Param("key"))
		wh.NodeServer.updateState(node.LogicChannel(), node)
		c.JSON(200, node.State())
		return
	}
	c.String(500, "node server is wrong type")
}

func (wh *WebHandler) GetServerSet(c *gin.Context) {
	if node, ok := wh.Nodes.ByName("server").(*servernode.Node); ok {
		node.Set(c.Param("key"), c.Param("value"))
		wh.NodeServer.updateState(node.LogicChannel(), node)
		c.JSON(200, node.State())
		return
	}
	c.String(500, "node server is wrong type")
}
