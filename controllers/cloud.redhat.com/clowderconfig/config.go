package clowderconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type ClowderConfig struct {
	Credentials struct {
		Keycloak struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
	}
	Images struct {
		MBOP     string `json:"mbop"`
		Caddy    string `json:"caddy"`
		Keycloak string `json:"Keycloak"`
	} `json:"images"`
	DebugOptions struct {
		Logging struct {
			DebugLogging bool `json:"debugLogging"`
		}
		Trigger struct {
			Diff bool `json:"diff"`
		} `json:"trigger"`
		Cache struct {
			Create bool `json:"create"`
			Update bool `json:"update"`
			Apply  bool `json:"apply"`
		} `json:"cache"`
		Pprof struct {
			Enable  bool   `json:"enable"`
			CpuFile string `json:"cpuFile"`
		} `json:"pprof"`
	} `json:"debugOptions"`
	Features struct {
		CreateServiceMonitor        bool `json:"createServiceMonitor"`
		DisableWebhooks             bool `json:"disableWebhooks"`
		WatchStrimziResources       bool `json:"watchStrimziResources"`
		UseComplexStrimziTopicNames bool `json:"useComplexStrimziTopicNames"`
		EnableAuthSidecarHook       bool `json:"enableAuthSidecarHook"`
		KedaResources               bool `json:"enableKedaResources"`
		PerProviderMetrics          bool `json:"perProviderMetrics"`
		ReconciliationMetrics       bool `json:"reconciliationMetrics"`
		DisableCloudWatchLogging    bool `json:"disableCloudWatchLogging"`
	} `json:"features"`
}

func getConfig() ClowderConfig {
	configPath := "/config/clowder_config.json"

	if path := os.Getenv("CLOWDER_CONFIG_PATH"); path != "" {
		configPath = path
	}

	fmt.Printf("Loading config from: %s\n", configPath)

	jsonData, err := ioutil.ReadFile(configPath)

	if err != nil {
		fmt.Printf("Config file not found\n")
		return ClowderConfig{}
	}

	clowderConfig := ClowderConfig{}
	err = json.Unmarshal(jsonData, &clowderConfig)

	if err != nil {
		fmt.Printf("Couldn't parse json:\n" + err.Error())
		return ClowderConfig{}
	}

	return clowderConfig
}

var LoadedConfig ClowderConfig

func init() {
	LoadedConfig = getConfig()
}
