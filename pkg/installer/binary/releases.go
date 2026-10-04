package binary

import (
	"context"
	"fmt"

	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
)

// GetLatestRelease returns the newest published release that has binaries attached.
func GetLatestRelease() (*github.RepositoryRelease, error) {
	client := github.NewClient(nil)
	releases, _, err := client.Repositories.ListReleases(context.Background(), "stampzilla", "stampzilla-go", &github.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error fetching releases from github.com: %w", err)
	}

	for _, r := range releases {
		if r.GetDraft() || r.GetPrerelease() {
			continue
		}
		for _, a := range r.Assets {
			if a.GetName() == "checksum" {
				return r, nil
			}
		}
		logrus.Warnf("skipping release %s: no binaries found", r.GetTagName())
	}

	return nil, fmt.Errorf("found no release with binaries on github.com")
}
