package main

import "github.com/gin-gonic/gin"

func initWebserver() {
	r := gin.Default()

	r.StaticFile("/", "./web/build/index.html")
	r.StaticFile("/manifest.json", "./web/build/manifest.json")
	r.Static("/static", "./web/build/static")
	r.Run(":8089")
}
