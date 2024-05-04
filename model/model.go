package model

import "time"

type ApplicationDetails struct {
	VersionID  string `json:"versionId"`
	Domain     string `json:"domain"`
	FullDomain string `json:"fullDomain"`
	Properties struct {
		AnypointPlatformConfigAnalyticsAgentEnabled string `json:"anypoint.platform.config.analytics.agent.enabled"`
		MuleEnv                                     string `json:"mule.env"`
		SecureKey                                   string `json:"secureKey"`
	} `json:"properties"`
	PropertiesOptions struct {
		SecureKey struct {
			Secure bool `json:"secure"`
		} `json:"secureKey"`
	} `json:"propertiesOptions"`
	Status  string `json:"status"`
	Workers struct {
		Type struct {
			Name   string  `json:"name"`
			Weight float64 `json:"weight"`
			CPU    string  `json:"cpu"`
			Memory string  `json:"memory"`
		} `json:"type"`
		Amount              int     `json:"amount"`
		RemainingOrgWorkers float64 `json:"remainingOrgWorkers"`
		TotalOrgWorkers     float64 `json:"totalOrgWorkers"`
	} `json:"workers"`
	LastUpdateTime int64  `json:"lastUpdateTime"`
	FileName       string `json:"fileName"`
	MuleVersion    struct {
		Version          string `json:"version"`
		UpdateID         string `json:"updateId"`
		LatestUpdateID   string `json:"latestUpdateId"`
		EndOfSupportDate int64  `json:"endOfSupportDate"`
	} `json:"muleVersion"`
	PreviousMuleVersion struct {
		Version          string `json:"version"`
		UpdateID         string `json:"updateId"`
		EndOfSupportDate int64  `json:"endOfSupportDate"`
	} `json:"previousMuleVersion"`
	Region                    string `json:"region"`
	MonitoringAutoRestart     bool   `json:"monitoringAutoRestart"`
	StaticIPsEnabled          bool   `json:"staticIPsEnabled"`
	HasFile                   bool   `json:"hasFile"`
	SecureDataGatewayEnabled  bool   `json:"secureDataGatewayEnabled"`
	LoggingNgEnabled          bool   `json:"loggingNgEnabled"`
	LoggingCustomLog4JEnabled bool   `json:"loggingCustomLog4JEnabled"`
	CloudObjectStoreRegion    string `json:"cloudObjectStoreRegion"`
	InsightsReplayDataRegion  string `json:"insightsReplayDataRegion"`
	IsDeploymentWaiting       bool   `json:"isDeploymentWaiting"`
	DeploymentGroup           struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"deploymentGroup"`
	UpdateRuntimeConfig bool `json:"updateRuntimeConfig"`
	TrackingSettings    struct {
		TrackingLevel string `json:"trackingLevel"`
	} `json:"trackingSettings"`
	LogLevels   []interface{} `json:"logLevels"`
	IPAddresses []interface{} `json:"ipAddresses"`
}

type Authorization struct {
	AccessToken string `json:"access_token"`
}

type AssetInformation struct {
	GroupID           string `json:"groupId"`
	AssetID           string `json:"assetId"`
	Version           string `json:"version"`
	Description       string `json:"description"`
	VersionGroup      string `json:"versionGroup"`
	ProductAPIVersion string `json:"productAPIVersion"`
	IsPublic          bool   `json:"isPublic"`
	Name              string `json:"name"`
	Type              string `json:"type"`
	IsSnapshot        bool   `json:"isSnapshot"`
	Status            string `json:"status"`
	AssetLink         string `json:"assetLink"`
}

type ListOfEnv struct {
	Data []struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		OrganizationID string `json:"organizationId"`
		IsProduction   bool   `json:"isProduction"`
		Type           string `json:"type"`
		ClientID       string `json:"clientId"`
	} `json:"data"`
	Total int `json:"total"`
}

type AssetDownload struct {
	Files []struct {
		Classifier   string      `json:"classifier"`
		Packaging    string      `json:"packaging"`
		ExternalLink string      `json:"externalLink"`
		Md5          string      `json:"md5"`
		Sha1         string      `json:"sha1"`
		CreatedDate  time.Time   `json:"createdDate"`
		MainFile     interface{} `json:"mainFile"`
		IsGenerated  bool        `json:"isGenerated"`
	} `json:"files"`
}

var ListofEnvId map[string]string = map[string]string{
	"SO2C-dev":  "",
	"SO2C-prod": "",
	"SO2C-sqa":  "",
	"SO2C-val":  "",
	//"Sandbox": "bff51d7c-a7d9-4781-b411-5a6fbaed40e1",
}

type AnypointPlatform struct {
	User struct {
		ContributorOfOrganizations []ChildEnv `json:"contributorOfOrganizations"`
	} `json:"user"`
}

type ChildEnv struct {
	Name       string `json:"name"`
	Id         string `json:"id"`
	ParentId   string `json:"parentId"`
	ParentName string `json:"parentName"`
	IsRoot     bool   `json:"isRoot"`
}
