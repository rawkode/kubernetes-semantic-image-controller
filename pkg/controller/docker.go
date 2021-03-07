package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver"
)

var (
	httpClient    = &http.Client{Timeout: 3 * time.Second}
	semanticRegex = regexp.MustCompile("^[=><].*")
)

// dockerHubResolver goes out to Docker Hub Registry to resolve image tags and
// figures out the best image to use
type dockerHubResolver struct{}

type dockerHubTag struct {
	Layer string `json:"layer"`
	Tag   string `json:"name"`
}

func (d *dockerHubResolver) Resolve(input string) (string, error) {
	imgName, constraint, ok := d.imageNameAndConstraint(input)
	if !ok {
		return "", fmt.Errorf("could not extract image name and version")
	}

	// Reach out to Docker Registry to fetch the list of images
	//
	// TODO: Docker Hub doesn't have an official API to get tags for an image
	// 		 This may stop working at any point or change in semantics
	rsp, err := httpClient.Get(fmt.Sprintf("https://registry.hub.docker.com/v1/repositories/%s/tags", imgName))
	if err != nil {
		return "", fmt.Errorf("could not fetch from Docker Hub: %v", err)
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("got bad status code: %d", rsp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %v", err)
	}

	// Parse out the images from Docker Hub
	tags := make([]dockerHubTag, 0)
	if err = json.Unmarshal(bytes, &tags); err != nil {
		return "", fmt.Errorf("could not unmarshal from Docker Hub: %v", err)
	}

	// Figure out which ones match our constraint
	candidates := make([]dockerHubTag, 0)
	for _, image := range tags {
		parsed, err := semver.NewVersion(image.Tag)
		if err != nil {
			continue
		}

		if constraint.Check(parsed) {
			candidates = append(candidates, image)
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("could not find image satisfying constraint")
	}

	// Figure out the most recent image that matches our constraints
	sort.Slice(candidates, func(i, j int) bool {
		pi, _ := semver.NewVersion(candidates[i].Tag)
		pj, _ := semver.NewVersion(candidates[j].Tag)
		return pi.LessThan(pj)
	})

	return fmt.Sprintf("%s:%s", imgName, candidates[len(candidates)-1].Tag), nil
}

func (d *dockerHubResolver) ShouldResolve(input string) bool {
	// We expect two components, separated by the colon
	split := strings.SplitN(input, ":", 2)
	if len(split) <= 1 {
		return false
	}

	// Check that it contains the semantic characters
	version := strings.TrimSpace(split[1])
	if !semanticRegex.MatchString(version) {
		return false
	}

	// If can be parsed, it's valid!
	_, err := semver.NewConstraint(version)
	return err == nil
}

func (d *dockerHubResolver) imageNameAndConstraint(input string) (string, *semver.Constraints, bool) {
	// We expect two components, separated by the colon
	split := strings.SplitN(input, ":", 2)
	if len(split) <= 1 {
		return "", nil, false
	}

	// Check that it contains the semantic characters
	imageName := strings.TrimSpace(split[0])
	version := strings.TrimSpace(split[1])
	if !semanticRegex.MatchString(version) {
		return "", nil, false
	}

	// If can be parsed, it's valid!
	constraint, err := semver.NewConstraint(version)
	if err != nil {
		return "", nil, false
	}

	return imageName, constraint, true
}
