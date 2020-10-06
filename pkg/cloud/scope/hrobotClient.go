package scope

import (
	"context"

	hrobot "github.com/nl2go/hrobot-go"
	"github.com/nl2go/hrobot-go/models"
)

// HrobotClient collects all methods used by the controller in the hrobot cloud API
type HrobotClient interface {
	UserName() string
	Password() string
	ResetBMServer(string, string) (*models.ResetPost, error)
	ListBMServers() ([]models.Server, error)
	ActivateRescue(string, string) (*models.Rescue, error)
	ListBMKeys() ([]models.Key, error)
	SetBMServerName(string, string) (*models.Server, error)
	GetBMServer(string) (*models.Server, error)
}

type HrobotClientFactory func(context.Context) (HrobotClient, error)

var _ HrobotClient = &realHrobotClient{}

type realHrobotClient struct {
	client   hrobot.RobotClient
	userName string
	password string
}

func (c *realHrobotClient) UserName() string {
	return c.userName
}

func (c *realHrobotClient) Password() string {
	return c.password
}

func (c *realHrobotClient) ResetBMServer(ip string, resetType string) (*models.ResetPost, error) {
	return c.client.ResetSet(ip, &models.ResetSetInput{Type: resetType})
}

func (c *realHrobotClient) ListBMServers() ([]models.Server, error) {
	return c.client.ServerGetList()
}

func (c *realHrobotClient) ActivateRescue(ip string, key string) (*models.Rescue, error) {
	return c.client.BootRescueSet(ip, &models.RescueSetInput{AuthorizedKey: key})
}

func (c *realHrobotClient) ListBMKeys() ([]models.Key, error) {
	return c.client.KeyGetList()
}

func (c *realHrobotClient) SetBMServerName(ip string, name string) (*models.Server, error) {
	return c.client.ServerSetName(ip, &models.ServerSetNameInput{Name: name})
}

func (c *realHrobotClient) GetBMServer(ip string) (*models.Server, error) {
	return c.client.ServerGet(ip)
}
