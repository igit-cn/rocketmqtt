package acl

import (
	"rocketmqtt/conf"

	"github.com/prometheus/common/log"
	"go.uber.org/zap"
)

type aclAuth struct {
	config *ACLConfig
	users  *map[string]string
}

func Init() *aclAuth {
	aclConfig, err := AclConfigLoad("conf/acl.conf")
	if err != nil {
		panic(err)
	}
	// content, err := ioutil.ReadFile("conf/mqttusers.json")
	// if err != nil {
	// 	log.Error("Read config file error: ", zap.Error(err))
	// 	panic(err)
	// }
	// users := new(map[string]string)
	// err = json.Unmarshal(content, &users)
	// if err != nil {
	// 	log.Error("Users config file unmarshal error: ", zap.Error(err))
	// 	panic(err)
	// }

	return &aclAuth{
		config: aclConfig,
		users:  &conf.RunConfig.Auth,
	}
}

func (a *aclAuth) CheckConnect(clientID, username, password string) bool {
	if (*a.users)[username] == "" {
		log.Warn("User not exist: ", zap.String("username", username), zap.String("password", password))
		return false
	}
	if (*a.users)[username] != password {
		log.Warn("User Authentication failed: ", zap.String("username", username), zap.String("password", password))
		return false
	}
	return true
}

func (a *aclAuth) CheckACL(action, clientID, username, ip, topic string) bool {
	return checkTopicAuth(a.config, action, username, ip, clientID, topic)
}
