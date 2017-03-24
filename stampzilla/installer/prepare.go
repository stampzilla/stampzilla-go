package installer

func Prepare() error {

	// Make sure our infrastructure is correct
	//// Create required user and folders
	CreateUser("stampzilla")
	CreateDirAsUser("/var/spool/stampzilla", "stampzilla")
	CreateDirAsUser("/var/log/stampzilla", "stampzilla")
	CreateDirAsUser("/home/stampzilla/go", "stampzilla")
	CreateDirAsUser("/etc/stampzilla", "stampzilla")
	CreateDirAsUser("/etc/stampzilla/nodes", "stampzilla")

	c := Config{}
	c.CreateConfig()

	return nil
}
