package helper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jchawla2804/golang-slack-event-listener/model"
)

const (
	Base_Url = "https://anypoint.mulesoft.com/"
)

// GetToken retrieves the token details based on the type of authentication.
// It takes the username, password, and typeOfAuth as input parameters.
// It returns the token details and an error if any.
func GetToken(username, password, typeOfAuth string) (interface{}, error) {
	log.Println("Retrieve Token Details")
	tokenDetails := model.Authorization{}
	var authurl string
	var body map[string]string

	switch typeOfAuth {
	case "basic-auth":
		body = map[string]string{
			"username": username,
			"password": password,
		}

		authurl = Base_Url + "accounts/login"

	case "oauth":
		body = map[string]string{
			"client_id":     username,
			"client_secret": password,
			"grant_type":    "client_credentials",
		}
		authurl = Base_Url + "accounts/api/v2/oauth2/token"

	default:
		return "", errors.New("invalid Auth Type")
	}

	log.Println(authurl)

	dataInBytes, err := json.Marshal(body)

	if err != nil {
		log.Print(err)
		return "", err
	}

	resp, err := http.Post(authurl, "application/json", bytes.NewBuffer(dataInBytes))

	if err != nil {
		log.Print(err.Error())
		return "", err

	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("status code is not correct")
		log.Println(resp.StatusCode)
		b, _ := io.ReadAll(resp.Body)
		log.Println(string(b))
		return "", err
	}

	log.Println(resp.StatusCode)
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&tokenDetails)

	return tokenDetails.AccessToken, nil

}

// GetPlatformInformation retrieves the platform information based on the token.
// It takes the token as an input parameter.
// It returns the platform details and an error if any.
func GetPlatformInformation(token string) (model.AnypointPlatform, error) {
	platformDetails := model.AnypointPlatform{}
	httpClient := &http.Client{}

	req, err := http.NewRequest("GET", Base_Url+"accounts/api/me", nil)
	if err != nil {
		log.Print(err.Error())
		return platformDetails, err
	}

	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Print(err.Error())
		return platformDetails, err
	}

	if resp.StatusCode != 200 {
		log.Printf("Status code is different %d", resp.StatusCode)
		return platformDetails, errors.New("error recieved while calling api")
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&platformDetails)

	return platformDetails, nil
}

// GetAppDetails retrieves the status of all applications deployed in a MuleSoft environment.
// It takes the token, envId, and orgId as input parameters.
// It returns the application details and an error if any.
func GetAppDetails(token string, envId, orgId string) (string, error) {
	appDetails := []model.ApplicationDetails{}
	httpClient := &http.Client{}

	req, err := http.NewRequest("GET", Base_Url+"/cloudhub/api/v2/applications", nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("X-ANYPNT-ORG-ID", orgId)
	req.Header.Add("X-ANYPNT-ENV-ID", envId)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println(err.Error())
		return "", err

	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("status code is not correct")
		log.Println(resp.StatusCode)
		b, _ := io.ReadAll(resp.Body)
		log.Println(string(b))
		return "", err
	}

	json.NewDecoder(resp.Body).Decode(&appDetails)

	var concatenatedString []string

	for _, value := range appDetails {
		concatenatedString = append(concatenatedString, "Name : "+value.Domain+"\n"+"Status : "+value.Status+"\n"+"Workers Cpu : "+value.Workers.Type.CPU+"\n")
	}

	return ("```" + strings.Join(concatenatedString, "\n\n") + "```"), nil

}

// ChangeAppStatus changes the status of an application (stop, start, restart).
// It takes the status, token, envId, orgId, and appName as input parameters.
// It returns a boolean indicating the success of the status change and an error if any.
func ChangeAppStatus(status string, token string, envId, orgId string, appName string) (bool, error) {
	httpClient := &http.Client{}
	body := map[string]string{
		"status": status,
	}

	dataInBytes, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", Base_Url+"cloudhub/api/applications/"+appName+"/status", bytes.NewBuffer(dataInBytes))
	if err != nil {
		log.Fatal(err)
		return false, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("X-ANYPNT-ORG-ID", orgId)
	req.Header.Add("X-ANYPNT-ENV-ID", envId)
	req.Header.Add("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println(err.Error())
		return false, err
	}

	defer resp.Body.Close()

	byte, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return false, fmt.Errorf("status update failed %s", string(byte))
	}

	return true, nil

}

// GetAssetInfo retrieves the information of all assets.
// It takes the token as an input parameter.
// It returns the asset details and an error if any.
func GetAssetInfo(token string) (string, error) {
	assetDetails := []model.AssetInformation{}
	httpClient := &http.Client{}

	req, err := http.NewRequest("GET", Base_Url+"exchange/api/v1/assets", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	q := req.URL.Query()
	q.Add("organizationId", os.Getenv("ANYPOINT_ORG_ID"))
	req.URL.RawQuery = q.Encode()
	log.Println(req.URL.String())

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println(err.Error())
		return "", err

	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("status code is not correct")
		log.Println(resp.StatusCode)
		b, _ := io.ReadAll(resp.Body)
		log.Println(string(b))
		return "", err
	}

	json.NewDecoder(resp.Body).Decode(&assetDetails)

	var concatenatedString []string

	for _, value := range assetDetails {
		concatenatedString = append(concatenatedString, fmt.Sprintf("Asset Name :- %s\n Group Id :- %s\n Asset Id :- %s\n Version :- %s\n Asset Link :- %s\n Description :- %s", value.Name, value.GroupID, value.AssetID, value.Version, value.AssetLink, value.Description))
	}

	return ("```" + strings.Join(concatenatedString, "\n\n") + "```"), nil

}

// ListEnvironments lists all the environments in a MuleSoft business group.
// It takes the token and orgId as input parameters.
// It returns the list of environments and an error if any.
func ListEnvironments(token, orgId string) (model.ListOfEnv, error) {
	envDetails := model.ListOfEnv{}
	httpClient := &http.Client{}

	req, err := http.NewRequest("GET", Base_Url+"accounts/api/organizations/"+orgId+"/environments", nil)
	if err != nil {
		log.Print(err.Error())
		return model.ListOfEnv{}, err
	}

	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Print(err.Error())
		return model.ListOfEnv{}, err
	}

	if resp.StatusCode != 200 {
		log.Printf("Status code is different %d", resp.StatusCode)
		return model.ListOfEnv{}, errors.New("error Recieved while calling api")
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&envDetails)

	return envDetails, nil
}

// DownloadAsset downloads an asset from Anypoint Exchange.
// It takes the token, orgid, and assetName as input parameters.
// It returns the downloaded asset file name and an error if any.
func DownloadAsset(token string, orgid, assetName string) (string, error) {
	specificAssetDetails := model.AssetDownload{}
	httpClient := &http.Client{}

	req, err := http.NewRequest("GET", Base_Url+"exchange/api/v1/assets/"+orgid+"/"+assetName, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}

	if resp.StatusCode != 200 {
		log.Printf("The Staus code is %d", resp.StatusCode)
		return "", errors.New("status code is note correct")
	}

	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&specificAssetDetails)
	assetLink := specificAssetDetails.Files[0].ExternalLink
	fileExtension := specificAssetDetails.Files[0].Packaging
	fileresp, err := http.Get(assetLink)
	if err != nil {
		log.Print(err.Error())
		return "", err
	}
	defer fileresp.Body.Close()
	out, err := os.Create(assetName + "." + fileExtension)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}

	defer out.Close()
	_, err = io.Copy(out, fileresp.Body)
	if err != nil {
		log.Print(err.Error())
		return "", err
	}

	return assetName + "." + fileExtension, nil

}
