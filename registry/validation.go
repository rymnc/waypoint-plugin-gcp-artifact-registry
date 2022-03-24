package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Location struct {
	Name       string            `json:"name"`
	Labels     map[string]string `json:"labels"`
	LocationId string            `json:"locationId"`
}

func getJson(url string, target interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func validateLocation(projectId string, location string) error {
	// fetching valid locations
	locations := new([]Location)

	urlWithProject := fmt.Sprintf("https://artifactregistry.googleapis.com/v1beta2/projects/%s/locations?alt=json", projectId)
	err := getJson(urlWithProject, locations)

	if err != nil {
		return err
	}

	//check if the location is one which can be used with the plugin
	locationValid := false
	validLocations := new([]string)

	for _, remoteLocation := range *locations {
		if remoteLocation.LocationId == location {
			locationValid = true
			*validLocations = append(*validLocations, remoteLocation.LocationId)
		}
	}

	// if location invalid, pick out locationId from all locations and print
	if locationValid == false {
		return fmt.Errorf("Invalid location '%s'. The valid locations are %s", location, strings.Join(*validLocations, ", "))
	}

	return nil
}
