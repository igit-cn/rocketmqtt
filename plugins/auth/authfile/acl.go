package acl

import (
	"rocketmqtt/conf"
	"rocketmqtt/logger"

	"go.uber.org/zap"
)

var log = logger.Instance.Named("acl")

type aclAuth struct {
	config *ACLConfig
	users  *map[string]string
}

func Init() *aclAuth {
	aclConfig, err := AclConfigLoad("conf/acl.conf")
	if err != nil {
		log.Panic(err.Error())
	}
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
