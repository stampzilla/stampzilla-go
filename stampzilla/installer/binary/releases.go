package binary

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Release struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Draft      bool   `json:"draft"`
	PreRelease bool   `json:"pre_release"`

	Date time.Time `json:"date"`

	Assets []Asset `json:"assets"`
}

type Asset struct {
	Name        string `json:"name"`
	Size        int    `json:"size"`
	DownloadURL string `json:"browser_download_url"`
}

func getReleases() []Release {
	url := "https://api.github.com/repos/stampzilla/stampzilla-go/releases"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return []Release{}
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return []Release{}
	}

	defer resp.Body.Close()

	var releases []Release

	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		log.Println(err)
	}

	return releases
}
