package main

import (
	"strings"

	"github.com/coreos/fleet/log"
	"github.com/stampzilla/stampzilla-go/stampzilla/installer"
)

func main() {
	//repoUrl := ""

	if NewCommitExists() {
		log.Info("New commit exists. Checking if we need to update")

		//GIT PULL

		// get the date git log --pretty=format:"%cI" -n 1 stampzilla-chromecast/

		// get the date of go/bin/stampzilla-chromecast

		//are they different? UPDATE!

	}

}

func NewCommitExists() bool {

	remoteSha1, err := getLatestServerCommit()
	if err != nil {
		log.Error(err)
		return false
	}

	localSha1, err := getLatestServerCommit()
	if err != nil {
		log.Error(err)
		return false
	}

	if remoteSha1 != localSha1 {
		return true
	}

	return false
}

func getLatestServerCommit() (sha1 string, err error) {
	result, err := installer.Run("git", "--git-dir", "/home/stampzilla/go/src/github.com/stampzilla/stampzilla-go/.git", "ls-remote", "origin", "master")
	sha1 = strings.Fields(result)[0]
	return
}

func getLatestLocalCommit() (sha1 string, err error) {
	result, err := installer.Run("git", "--git-dir", "/home/stampzilla/go/src/github.com/stampzilla/stampzilla-go/.git", "rev-parse", "HEAD")
	sha1 = strings.Fields(result)[0]
	return
}
