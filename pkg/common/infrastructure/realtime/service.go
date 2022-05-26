package realtime

import (
	"github.com/centrifugal/centrifuge-go"
	"github.com/pkg/errors"
)

type Message struct {
	ChannelID string
	Data      []byte
}

type ClientService interface {
	GetHost(id string) string
}

type Service interface {
	Publish(messages []Message) error
	Close() error
}

type clientService struct {
	hosts []string
}

type service struct {
	subscriptions []*centrifuge.Subscription
	clients       []*centrifuge.Client
	channelPrefix string
}

func NewClientService(hosts []string) ClientService {
	return &clientService{hosts: hosts}
}

func (s *clientService) GetHost(id string) string {
	return s.getHost(id, s.hosts)
}

func (s *clientService) getHost(id string, hosts []string) string {
	if len(hosts) == 0 {
		return ""
	}
	bytes := []byte(id)
	sum := 0
	for _, item := range bytes {
		sum += int(item)
	}
	return hosts[sum%len(hosts)]
}

func NewService(hosts []string, channelPrefix string) (Service, error) {
	if len(hosts) == 0 {
		return nil, nil
	}
	subscriptions := []*centrifuge.Subscription{}
	clients := []*centrifuge.Client{}
	for _, host := range hosts {
		client := centrifuge.NewJsonClient(host, centrifuge.DefaultConfig())
		err := client.Connect()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		subscription, err := client.NewSubscription(channelPrefix)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		err = subscription.Subscribe()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		subscriptions = append(subscriptions, subscription)
		clients = append(clients, client)
	}
	return &service{subscriptions: subscriptions, clients: clients, channelPrefix: channelPrefix}, nil
}

func (s *service) Close() error {
	for _, client := range s.clients {
		err := client.Close()
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (s *service) Publish(messages []Message) error {
	for _, message := range messages {
		subscription := s.getSubscription(message.ChannelID, s.subscriptions)
		client := s.getClient(message.ChannelID)
		if client == nil {
			continue
		}
		if subscription == nil {
			continue
		}
		//_, err := subscription.Publish(message.Data)
		_, err := client.Publish(s.channelPrefix+":"+message.ChannelID, message.Data)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (s *service) getClient(id string) *centrifuge.Client {
	if len(s.clients) == 0 {
		return nil
	}
	bytes := []byte(id)
	sum := 0
	for _, item := range bytes {
		sum += int(item)
	}
	return s.clients[sum%len(s.clients)]
}

func (s *service) getSubscription(id string, subscriptions []*centrifuge.Subscription) *centrifuge.Subscription {
	if len(subscriptions) == 0 {
		return nil
	}
	bytes := []byte(id)
	sum := 0
	for _, item := range bytes {
		sum += int(item)
	}
	return subscriptions[sum%len(subscriptions)]
}
