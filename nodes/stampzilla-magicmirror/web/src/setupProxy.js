const proxy = require("http-proxy-middleware")

module.exports = app => {
  app.use(proxy("/ws", {target: "http://localhost:8089", ws: true}))
}
