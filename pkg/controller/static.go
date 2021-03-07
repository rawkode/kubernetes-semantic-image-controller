package controller

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
)

// staticResolver is used to resolve in-memory version lists
type staticResolver struct {
	images map[string][]string
}

func (r *staticResolver) Resolve(input string) (string, error) {
	imgName, constraint, ok := imageNameAndConstraint(input)
	if !ok {
		return "", fmt.Errorf("could not extract image name and version")
	}

	imageVersions, ok := r.images[imgName]
	if !ok {
		return "", fmt.Errorf("no versions for image: %s", imgName)
	}

	// Figure out which ones match our constraint
	candidates := make([]string, 0)
	for _, version := range imageVersions {
		parsed, err := semver.NewVersion(version)
		if err != nil {
			continue
		}

		if constraint.Check(parsed) {
			candidates = append(candidates, version)
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("could not find image satisfying constraint")
	}

	// Figure out the most recent image that matches our constraints
	sort.Slice(candidates, func(i, j int) bool {
		pi, _ := semver.NewVersion(candidates[i])
		pj, _ := semver.NewVersion(candidates[j])
		return pi.LessThan(pj)
	})

	return fmt.Sprintf("%s:%s", imgName, candidates[len(candidates)-1]), nil
}

func (r *staticResolver) ShouldResolve(input string) bool {
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

func imageNameAndConstraint(input string) (string, *semver.Constraints, bool) {
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
