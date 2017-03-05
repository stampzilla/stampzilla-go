package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/go-github/github"
	"github.com/stampzilla/stampzilla-go/stampzilla/installer"
)

const (
	gitPath = "/home/stampzilla/go/src/github.com/stampzilla/stampzilla-go"
)

func main() {
	//installer := installer.NewInstaller()
	//installer.GoGet("github.com/stampzilla/stampzilla-go/nodes/stampzilla-enocean", true)
	//return

	wg := &sync.WaitGroup{}
	quit := make(chan struct{})
	go signalHandler(wg, quit)
	go update(wg)

	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				go update(wg)
			case <-quit:
				wg.Wait()
				ticker.Stop()
				return
			}
		}
	}()

	select {
	case <-quit:
		return
	}
}

func update(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	if NewCommitExists() {
		log.Println("New commit exists. Checking if we need to update")
		installer := installer.NewInstaller()
		nodes, err := ioutil.ReadDir("/home/stampzilla/go/bin")
		if err != nil {
			log.Println(err)
			return
		}
		for _, n := range nodes {
			if !strings.Contains(n.Name(), "stampzilla-") {
				continue
			}
			repoDate, err := getLatestCommitTimeInFolder(path.Join("nodes", n.Name()))
			if err != nil {
				log.Println(err)
				return
			}

			if repoDate.After(n.ModTime()) {
				log.Printf("Update date: %s for node: %s\n", repoDate, n.Name())
				installer.GoGet("github.com/stampzilla/stampzilla-go/nodes/"+n.Name(), true)
			}
		}
	}
}

func getLatestCommitTimeInFolder(folder string) (*time.Time, error) {
	client := github.NewClient(nil)
	ctx := context.Background()
	commits, _, err := client.Repositories.ListCommits(ctx, "stampzilla", "stampzilla-go",
		&github.CommitsListOptions{
			Path: folder,
			SHA:  "master",
		})

	if err != nil {
		return nil, err
	}
	return commits[0].Commit.Committer.Date, nil
}

func NewCommitExists() bool {
	remoteSha1, err := getLatestServerCommit()
	if err != nil {
		log.Println(err)
		return false
	}
	localSha1, err := getLatestLocalCommit()
	if err != nil {
		log.Println(err)
		return false
	}
	if remoteSha1 != localSha1 {
		return true
	}
	return false
}

func getLatestServerCommit() (sha1 string, err error) {
	result, err := installer.Run("git", "--git-dir", path.Join(gitPath, ".git"), "ls-remote", "origin", "master")
	sha1 = strings.Fields(result)[0]
	return
}

func getLatestLocalCommit() (sha1 string, err error) {
	result, err := installer.Run("git", "--git-dir", path.Join(gitPath, ".git"), "rev-parse", "HEAD")
	sha1 = strings.Fields(result)[0]
	return
}
func signalHandler(wg *sync.WaitGroup, quit chan struct{}) {
	ch := make(chan os.Signal, 10)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	for {
		<-ch
		log.Println("Shutdown requested. Finishing what we where doing...")
		wg.Wait()
		log.Println("Shutting down")
		close(quit)
		return
	}
}
