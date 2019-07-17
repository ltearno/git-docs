package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type AccessTokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
}

type UserCreationResponse struct {
	UUID             string              `json:"uuid"`
	ID               string              `json:"id"`
	Roles            map[string][]string `json:"roles"`
	Identityprovider interface{}         `json:"identityprovider"`
	Suid             interface{}         `json:"suid"`
}

type ApiKeyCreationRequest struct {
	Description string `json:"description"`
}

type ApiKeyCreationResponse struct {
	ID             string    `json:"id"`
	Description    string    `json:"description"`
	Secret         string    `json:"secret"`
	ExpirationDate time.Time `json:"expirationDate"`
	CreationDate   time.Time `json:"creationDate"`
	Date           time.Time `json:"date"`
	Loa            int       `json:"loa"`
	CreatedBy      string    `json:"createdBy"`
}

const MasterRealm = "IDP"
const AuthenticateServer = "authenticate-dev.idp.private.geoapi-airbusds.com"
const AuthorizeServer = "authorize-dev.idp.private.geoapi-airbusds.com"
const InitialApiKey = "eVjnw3Fo7bM6Q5LWNFnYWbnJlLfH814omUnlLLQoZn4KmAUE4uSOQRB9ZsI-kZuc2glH7fqMfvcAKx4P6W7RtA=="

func fetchToken(clientID string, apiKey string) (*string, interface{}) {
	tokenURL := fmt.Sprintf("https://%s/auth/realms/%s/protocol/openid-connect/token", AuthenticateServer, MasterRealm)
	req, err := http.NewRequest(http.MethodPost,
		tokenURL,
		strings.NewReader(fmt.Sprintf("grant_type=api_key&client_id=%s&apikey=%s",
			url.QueryEscape(clientID),
			url.QueryEscape(InitialApiKey))))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Sprintf("Bad response code, body=%s", string(body))
	}

	var record AccessTokenResponse

	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		return nil, err
	}

	return &record.AccessToken, nil
}

func createUser(email string, roles map[string][]string, accessToken string) (*UserCreationResponse, interface{}) {
	rolesJSON, err := json.Marshal(roles)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://%s/api/v1/users", AuthorizeServer), strings.NewReader(fmt.Sprintf("id=%s&roles=%s", url.QueryEscape(email), url.QueryEscape(string(rolesJSON)))))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Sprintf("Bad response code, body=%s", string(body))
	}

	var record UserCreationResponse

	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		return nil, err
	}

	return &record, nil
}

func createApiKey(userUUID string, apiKeyDescription string, accessToken string) (*ApiKeyCreationResponse, interface{}) {
	requestBody := ApiKeyCreationRequest{
		Description: apiKeyDescription,
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://%s/api/v1/apikeys", AuthorizeServer), strings.NewReader(string(requestBodyJSON)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("X-Forwarded-User", userUUID)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Sprintf("Bad response code %d, body=%s", resp.StatusCode, string(body))
	}

	var record ApiKeyCreationResponse

	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		return nil, err
	}

	return &record, nil
}

func main() {
	token, err := fetchToken("AAA", InitialApiKey)
	if err != nil {
		fmt.Println("Cannot get a token with the initial root api key")
		panic(err)
	}

	fmt.Printf("Got access token, creating a user...\n")

	roles := map[string][]string{
		"geo.idp.authorize":   []string{"admin", "geo.idp.impersonation"},
		"geo.gds.authorize":   []string{"admin", "geo.gds.impersonation"},
		"geo.theos.authorize": []string{"admin", "geo.theos.impersonation"},
	}

	user, err := createUser(fmt.Sprintf("dssskjfhlsdkjfhsfsdf%d@gmail.com", time.Now().Unix()), roles, *token)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Ok, user %s created\n", user.UUID)

	apiKey, err := createApiKey(user.UUID, "Installation process", *token)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created API Key id %s, secret %s\n", apiKey.ID, apiKey.Secret)

	/**
	TODO :
	- create same user in Subscription Management
	- associate it with the same rights
	*/
}
