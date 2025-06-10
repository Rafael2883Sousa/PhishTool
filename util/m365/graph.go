package m365

import (
	"encoding/json"
	"net/http"

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
