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

func GetToken(username, password string) (interface{}, error) {
	log.Println("Retrieve Token Details")
	tokenDetails := model.Authorization{}
	body := map[string]string{
		"username": username,
		"password": password,
	}

	dataInBytes, err := json.Marshal(body)

	if err != nil {
		log.Print(err)
		return "", err
	}

	resp, err := http.Post("https://anypoint.mulesoft.com/accounts/login", "application/json", bytes.NewBuffer(dataInBytes))

	if err != nil {
		log.Print(err.Error())
		return "", err

	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("Status code is not correct")
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

func GetPlatformInformation(token string) (model.AnypointPlatform, error) {
	platformDetails := model.AnypointPlatform{}
	httpClient := &http.Client{}

	req, err := http.NewRequest("GET", "https://anypoint.mulesoft.com/accounts/api/me", nil)
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
		return platformDetails, errors.New("error Recieved while calling api")
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&platformDetails)

	return platformDetails, nil
}

/*
Get Status of all applications deployed in a mulesoft environment
*/
func GetAppDetails(token string, envName string) (string, error) {
	appDetails := []model.ApplicationDetails{}
	httpClient := &http.Client{}

	req, err := http.NewRequest("GET", "https://anypoint.mulesoft.com/cloudhub/api/v2/applications", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("X-ANYPNT-ORG-ID", os.Getenv("ANYPOINT_ORG_ID"))
	req.Header.Add("X-ANYPNT-ENV-ID", model.ListofEnvId[envName])

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println(err.Error())
		return "", err

	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("Status code is not correct")
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

/*
Change app status whether to stop, start or restart
*/
func ChangeAppStatus(status string, token string, envName string, appName string) (bool, error) {
	httpClient := &http.Client{}
	body := map[string]string{
		"status": status,
	}

	dataInBytes, err := json.Marshal(body)

	req, err := http.NewRequest("POST", "https://anypoint.mulesoft.com/cloudhub/api/applications/"+appName+"/status", bytes.NewBuffer(dataInBytes))
	if err != nil {
		log.Fatal(err)
		return false, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("X-ANYPNT-ORG-ID", os.Getenv("ANYPOINT_ORG_ID"))
	req.Header.Add("X-ANYPNT-ENV-ID", model.ListofEnvId[envName])
	req.Header.Add("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println(err.Error())
		return false, err
	}

	defer resp.Body.Close()

	byte, _ := io.ReadAll(resp.Body)
	log.Printf(string(byte))
	if resp.StatusCode != 200 {
		return false, fmt.Errorf("It Failed to Update the status")
	}

	return true, nil

}

/*
Get List of all assets
*/
func GetAssetInfo(token string) (string, error) {
	assetDetails := []model.AssetInformation{}
	httpClient := &http.Client{}

	req, err := http.NewRequest("GET", "https://anypoint.mulesoft.com/exchange/api/v1/assets", nil)
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
		err = fmt.Errorf("Status code is not correct")
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

/*
List all the environments in mulesoft business group
*/
func ListEnvironments(token string) (string, error) {
	envDetails := model.ListOfEnv{}
	httpClient := &http.Client{}

	req, err := http.NewRequest("GET", "https://anypoint.mulesoft.com/accounts/api/organizations/"+os.Getenv("ANYPOINT_ORG_ID")+"/environments", nil)
	if err != nil {
		log.Print(err.Error())
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Print(err.Error())
		return "", err
	}

	if resp.StatusCode != 200 {
		log.Printf("Status code is different %d", resp.StatusCode)
		return "", errors.New("Error Recieved while calling api")
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&envDetails)

	var concatenatedString []string
	for _, v := range envDetails.Data {
		concatenatedString = append(concatenatedString, fmt.Sprintf("Env-Name : %s\n Env-Id : %s\n Is-Production : %v", v.Name, v.ID, v.IsProduction))
	}

	return strings.Join(concatenatedString, "\n\n"), nil
}

/*
Download any asset from anypoint exchange
*/
func DownloadAsset(token string, assetName string) (string, error) {
	specificAssetDetails := model.AssetDownload{}
	httpClient := &http.Client{}

	req, err := http.NewRequest("GET", "https://anypoint.mulesoft.com/exchange/api/v1/assets/"+os.Getenv("ANYPOINT_ORG_ID")+"/"+assetName, nil)
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
		return "", errors.New("Status code is note correct")
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
		log.Printf(err.Error())
		return "", err
	}

	return assetName + "." + fileExtension, nil

}
