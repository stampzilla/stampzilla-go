package binary

import (
	"context"

	"github.com/google/go-github/github"
)

func GetReleases() []*github.RepositoryRelease {
	client := github.NewClient(nil)
	ctx := context.Background()
	releases, _, err := client.Repositories.ListReleases(ctx, "stampzilla", "stampzilla-go", &github.ListOptions{})
	if err != nil {
		return nil
	}

	return releases

	////commits, _, err := client.Repositories.ListCommits(ctx, "stampzilla", "stampzilla-go"

	// url := "https://api.github.com/repos/stampzilla/stampzilla-go/releases"

	//req, err := http.NewRequest("GET", url, nil)
	//if err != nil {
	//log.Fatal("NewRequest: ", err)
	//return []Release{}
	//}

	// client := &http.Client{}

	//resp, err := client.Do(req)
	//if err != nil {
	//log.Fatal("Do: ", err)
	//return []Release{}
	//}

	// defer resp.Body.Close()

	// var releases []Release

	//if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
	//log.Println(err)
	//}

	// return releases
}
