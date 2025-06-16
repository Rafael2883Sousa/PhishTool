package m365

import (
	"encoding/json"
	"net/http"
	"fmt"
	"io/ioutil"

	"github.com/gophish/gophish/models"
)

func FetchGroupsFromGraph(token string) ([]models.Group, error) {
	req, _ := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/groups", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Value []struct {
			DisplayName string `json:"displayName"`
			ID          string `json:"id"`
		} `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var groups []models.Group
	for _, g := range result.Value {
		groups = append(groups, models.Group{
			Name: g.DisplayName,
		})
	}
	return groups, nil
}

func FetchUsersFromGraph(token string) ([]models.Target, error) {
	url := "https://graph.microsoft.com/v1.0/users?$select=displayName,givenName,surname,mail,jobTitle"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Graph API returned status %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Estrutura de resposta do Graph API
	var result struct {
		Value []struct {
			Mail       string `json:"mail"`
			GivenName  string `json:"givenName"`
			Surname    string `json:"surname"`
			JobTitle   string `json:"jobTitle"`
		} `json:"value"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	// Mapear para models.Target
	var targets []models.Target
	for _, user := range result.Value {
		if user.Mail == "" {
			continue // ignorar sem e-mail
		}
		targets = append(targets, models.Target{
			BaseRecipient: models.BaseRecipient{
				Email:     user.Mail,
				FirstName: user.GivenName,
				LastName:  user.Surname,
				Position:  user.JobTitle,
			},
		})
	}

	return targets, nil
}
