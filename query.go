package syntropy

import (
	"encoding/json"
	"fmt"
	"github.com/imkira/go-ttlmap"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type SyntropyToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type SyntropyAgentService struct {
	ServiceId int                          `json:"agent_service_id"`
	AgentId   int                          `json:"agent_id"`
	Name      string                       `json:"agent_service_name"`
	Subnets   []SyntropyAgentServiceSubnet `json:"agent_service_subnets"`
	Active    bool                         `json:"agent_service_is_active"`
}

type SyntropyAgentServiceSubnet struct {
	Ip string `json:"agent_service_subnet_ip"`
}

type SyntropyAgentServicesResponse struct {
	Data []SyntropyAgentService `json:"data"`
}

type SyntropyAgent struct {
	Id     int    `json:"agent_id"`
	Name   string `json:"agent_name"`
	Active bool   `json:"agent_is_online"`
}

type SyntropyAgentResponse struct {
	Data []SyntropyAgent `json:"data"`
}

var localCache = ttlmap.New(nil)

func login(controller_url string, username string, password string) string {
	token := SyntropyToken{}

	data := url.Values{
		"user_email":    {username},
		"user_password": {password},
	}

	resp, err := http.PostForm(fmt.Sprintf("%s/api/auth/local/login", controller_url), data)

	if err != nil {
		log.Fatalf("Login HTTP error %v", err)
		return ""
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Login error status code: %v", resp.StatusCode)
		return ""
	}

	body, _ := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatalf("Failed to decode token data %v", err)
	}

	jsonAns := string(body)
	err = json.Unmarshal([]byte(jsonAns), &token)

	if err != nil {
		log.Fatalf("Failed to decode token data %v", err)
	}

	return token.AccessToken
}

func query(dns_name string, controller_url string, token string, ttl time.Duration) string {
	item, err := localCache.Get(dns_name)
	if err == nil {
		return item.Value().(string)
	}

	service_ips := get_service_ips(controller_url, token, ttl)

	ip, ok := service_ips[dns_name]

	log.Info(dns_name)
	log.Info(service_ips)

	localCache.Set(dns_name, ttlmap.NewItem(ip, ttlmap.WithTTL(ttl)), nil)

	if !ok {
		return ""
	}

	return ip
}

func get_service_ips(controller_url string, token string, ttl time.Duration) map[string]string {
	agents := get_agents(controller_url, token)
	service_ips := make(map[string]string)

	client := &http.Client{}

	for id, agent := range agents {
		log.Info(fmt.Sprintf("%s/api/platform/agent-services?agent-id=%v", controller_url, id))
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/platform/agent-services?agent-ids=%v", controller_url, id), nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := client.Do(req)

		if err != nil {
			log.Fatalf("GET Agent services HTTP error %v", err)
			return nil
		}

		if resp.StatusCode != http.StatusOK {
			log.Fatalf("GET Agent services failure status code %v", resp.StatusCode)
			return nil
		}

		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Fatalf("Failed to decode agent services data %v", err)
		}

		syntropyResp := SyntropyAgentServicesResponse{}
		jsonAns := string(body)
		err = json.Unmarshal([]byte(jsonAns), &syntropyResp)

		for _, service := range syntropyResp.Data {
			if !service.Active {
				continue
			}
			service_ips[fmt.Sprintf("%s.%s", service.Name, agent.Name)] = service.Subnets[0].Ip
		}
	}

	return service_ips

}

func get_agents(controller_url string, token string) map[int]SyntropyAgent {
	client := &http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/platform/agents", controller_url), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("GET Agents HTTP error %v", err)
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("GET Agents failure status code %v", resp.StatusCode)
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatalf("Failed to decode agents data %v", err)
	}

	syntropyResp := SyntropyAgentResponse{}
	jsonAns := string(body)
	err = json.Unmarshal([]byte(jsonAns), &syntropyResp)

	if err != nil {
		log.Fatalf("Failed to decode agents data %v", err)
	}

	agents := make(map[int]SyntropyAgent)

	for _, agent := range syntropyResp.Data {
		if !agent.Active {
			continue
		}
		agents[agent.Id] = agent
	}

	return agents
}
