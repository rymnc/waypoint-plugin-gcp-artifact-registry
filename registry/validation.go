package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/go-playground/validator"
)

type Locations struct {
	Locations []Location `json:"locations"`
}

type Location struct {
	Name       string            `json:"name"`
	Labels     map[string]string `json:"labels"`
	LocationId string            `json:"locationId"`
}

type GCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type LocationError struct {
	Error GCPError `json:"error"`
}

func getLocationJson(url string, target interface{}) error {
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to get json from url: %s", url)
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("failed to get json from url: %s, status code: %d", url, res.StatusCode)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)

	if readErr != nil {
		return fmt.Errorf("failed to read response body: %s", readErr.Error())
	}

	errorResponse := new(LocationError)
	err = json.Unmarshal(body, target)

	if err != nil {
		err = json.Unmarshal(body, errorResponse)
		if err != nil {
			return fmt.Errorf("failed to parse response from GCP: %s", err.Error())
		}
		return fmt.Errorf("%s", errorResponse.Error.Message)
	}

	return nil
}

func validateLocation(projectId string, location string) error {
	// fetching valid locations
	locations := new(Locations)

	urlWithProject := fmt.Sprintf("https://artifactregistry.googleapis.com/v1beta2/projects/%s/locations?alt=json", projectId)
	err := getLocationJson(urlWithProject, locations)

	if err != nil {
		return fmt.Errorf("Failed to get locations from GCP: %s", err.Error())
	}

	//check if the location is one which can be used with the plugin
	locationValid := false
	validLocations := []string{}

	for _, remoteLocation := range *&locations.Locations {
		if remoteLocation.LocationId == location {
			locationValid = true
		}
		validLocations = append(validLocations, remoteLocation.LocationId)
	}

	if len(validLocations) == 0 {
		// this means that the project id is invalid
	}

	// if location invalid, pick out locationId from all locations and print
	if locationValid == false {
		return fmt.Errorf("Invalid location '%s'. The valid locations are %s", location, strings.Trim(strings.Join(validLocations, " "), " "))
	}

	return nil
}

var ErrInvalidRepositoryId = fmt.Errorf("Invalid repository id")
var ErrInvalidLocation = fmt.Errorf("Invalid location")

// validates the configuration passed to the plugin
func validateConfig(c RegistryConfig) error {
	v := validator.New()

	err := v.Struct(c)

	if err != nil {
		errorMessage := ""
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Namespace() {
			case "Config.Location":
				errorMessage += ErrInvalidLocation.Error()
			case "Config.RepositoryId":
				errorMessage += ErrInvalidRepositoryId.Error()
			default:
				errorMessage += fmt.Sprintf("%s\n", err.Value())
			}
		}

		// if
		return fmt.Errorf(errorMessage)
	}

	return nil
}
