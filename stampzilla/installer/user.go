package installer

import (
	"os"
	"os/user"
	"strconv"

	"github.com/Sirupsen/logrus"
)

func CreateUser(username string) {
	action := "Checking user '" + username + "'... "
	if userExists(username) {
		logrus.Debug(action + "(exists) DONE")
		return
	}

	out, err := Run("useradd", "-m", "-r", "-s", "/bin/false", username)
	if err != nil {
		logrus.Error(action+"ERROR", err, out)
		return
	}

	logrus.Info(action + "(created) DONE")
}
func userExists(username string) bool {
	_, err := Run("id", "-u", username)
	if err != nil {
		return false
	}
	return true
}

func CreateDirAsUser(directory string, username string) {
	action := "Check directory " + directory + "... "

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err := os.MkdirAll(directory, 0777)
		if err != nil {
			logrus.Error(action+"ERROR", err)
			return
		}
		logrus.Info(action + "(created) ")
	} else {
		action += "(exists) "
	}

	u, err := user.Lookup(username)
	if err != nil {
		logrus.Error(action+"ERROR: User lookup", err)
		return
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		logrus.Error(action+"ERROR: strconv.Atoi(u.Uid)", err)
		return
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		logrus.Error(action+"ERROR: strconv.Atoi(u.Gid)", err)
		return
	}
	err = os.Chown(directory, uid, gid)
	if err != nil {
		logrus.Error(action+"ERROR: Set permissions", err)
		return
	}

	logrus.Debug(action + "DONE")
}
