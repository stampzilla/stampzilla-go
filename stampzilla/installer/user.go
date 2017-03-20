package installer

import (
	"os"
	"os/user"
	"strconv"

	"github.com/Sirupsen/logrus"
)

func CreateUser(username string) {
	action := "Check user '" + username + "'"
	if userExists(username) {
		logrus.WithFields(logrus.Fields{
			"exists": "true",
		}).Debug(action)
		return
	}

	out, err := Run("useradd", "-m", "-r", "-s", "/bin/false", username)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"exists": "false",
			"error":  err,
			"output": out,
		}).Panic(action)
		return
	}

	logrus.WithFields(logrus.Fields{
		"exists":  "false",
		"created": "true",
	}).Info(action)
}
func userExists(username string) bool {
	_, err := Run("id", "-u", username)
	if err != nil {
		return false
	}
	return true
}

func CreateDirAsUser(directory string, username string) {
	action := "Check directory " + directory
	var created = false

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err := os.MkdirAll(directory, 0777)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"exists": "false",
				"error":  err,
			}).Panic(action)
			return
		}
		created = true
	}

	u, err := user.Lookup(username)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"exists": "true",
			"step":   "User lookup",
			"error":  err,
		}).Panic(action)
		return
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"exists": "true",
			"step":   "strconv.Atoi(u.Uid)",
			"error":  err,
		}).Panic(action)
		return
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"exists": "true",
			"step":   "strconv.Atoi(u.Gid)",
			"error":  err,
		}).Panic(action)
		return
	}
	err = os.Chown(directory, uid, gid)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"exists": "true",
			"step":   "Set permissions",
			"error":  err,
		}).Panic(action)
		return
	}

	if created {
		logrus.WithFields(logrus.Fields{
			"created": "true",
		}).Info(action)
	} else {
		logrus.WithFields(logrus.Fields{
			"exists": "true",
		}).Debug(action)
	}
}
