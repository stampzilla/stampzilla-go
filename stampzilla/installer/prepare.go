package installer

func Prepare() error {

	//nodes, err := ioutil.ReadDir("/home/stampzilla/go/src/github.com/stampzilla/stampzilla-go/nodes/")
	//if err != nil {
	//fmt.Println("Found no nodes. installing stampzilla cli first!")
	//CreateUser("stampzilla")
	//CreateDirAsUser("/home/stampzilla/go", "stampzilla")
	//GoGet("github.com/stampzilla/stampzilla-go/stampzilla", upgrade)
	//}

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
