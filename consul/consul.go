package consul

import (
	capi "github.com/hashicorp/consul/api"
)

func GetClientFromConfig(c *capi.Config) *capi.Client {
	client, err := capi.NewClient(c)
	if err != nil {
		panic(err)
	}
	return client
}
func GetClient() *capi.Client {
	client, err := capi.NewClient(capi.DefaultConfig())
	if err != nil {
		panic(err)
	}
	return client
}

func RegisterService(c *capi.Client) {
	s := c.Agent()
	serviceDefinition := capi.AgentServiceRegistration{
		Name:              "",
		Address:           "",
		SocketPath:        "",
		Check:             &capi.AgentServiceCheck{},
	}
    // TODO Add Service Struct
	s.ServiceRegister(&serviceDefinition)
}

func DeregisterService() {

}
