package broker

import (
	"github.com/gin-gonic/gin"
	"rocketmqtt/conf"
)

func InitHTTPMoniter(b *Broker) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.DELETE("api/v1/connections/:clientid", func(c *gin.Context) {
		clientid := c.Param("clientid")
		cli, ok := b.clients.Load(clientid)
		if ok {
			conn, succss := cli.(*client)
			if succss {
				conn.Close()
			}
		}
		resp := map[string]int{
			"code": 0,
		}
		c.JSON(200, &resp)
	})
	router.GET("api/v1/connections/:clientid", func(c *gin.Context) {
		clientid := c.Param("clientid")
		cli, ok := b.clients.Load(clientid)
		if ok {
			conn, succss := cli.(*client)
			if succss {
				topics, qos, _ := conn.session.Topics()
				c.JSON(200, map[string]interface{}{
					"clientID":  conn.info.clientID,
					"username":  conn.info.username,
					"localIP":   conn.info.localIP,
					"remoteIP":  conn.info.remoteIP,
					"keepalive": conn.info.keepalive,
					"status":    conn.status,
					"topics":    topics,
					"qos":       qos,
				})
				return
			}
		}
		c.JSON(200, map[string]string{
			"msg": "client not found",
		})
	})
	router.GET("api/v1/sessions/:clientid", func(c *gin.Context) {
		clientid := c.Param("clientid")
		s, err := b.sessionMgr.Get(clientid)
		if err == nil {
			c.JSON(200, map[string]string{
				"ID": s.ID(),
			})
			return
		}
		c.JSON(200, map[string]string{
			"msg": err.Error(),
		})
	})
	router.GET("api/v1/connections", func(c *gin.Context) {
		var clients []string
		b.clients.Range(func(key, value interface{}) bool {
			clients = append(clients, key.(string))
			return true
		})
		c.JSON(200, map[string]interface{}{
			"count":   len(clients),
			"clients": clients,
		})
	})
	router.Run(":" + conf.RunConfig.HTTPPort)
}
