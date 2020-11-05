package nx

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

type NxSender struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func New(parameters json.RawMessage) *NxSender {
	nx := &NxSender{}

	json.Unmarshal(parameters, nx)

	return nx
}

func (nx *NxSender) Trigger(dest []string, body string) error {
	return nx.notify(true, dest, body)
}

func (nx *NxSender) Release(dest []string, body string) error {
	return nx.notify(false, dest, body)
}

func (nx *NxSender) notify(trigger bool, dest []string, body string) error {
	u, err := url.Parse(nx.Server)
	if err != nil {
		return err
	}

	values := map[string][]string{
		"cameraRefs": dest,
	}
	metadata, _ := json.Marshal(values)

	q := u.Query()

	q.Set("source", "stampzilla")
	q.Set("metadata", string(metadata))
	q.Set("caption", body)

	if trigger {
		q.Set("state", "Active")
	} else {
		q.Set("state", "Inactive")
	}

	u.Path = "/api/createEvent"
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(nx.Username, nx.Password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// b, err := ioutil.ReadAll(resp.Body)
	// spew.Dump(b)

	return err
}

func (nx *NxSender) Destinations() (map[string]string, error) {
	u, err := url.Parse(nx.Server)
	if err != nil {
		return nil, err
	}

	u.Path = "/ec2/getCamerasEx"

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(nx.Username, nx.Password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	type Response []struct {
		AddParams []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"addParams"`
		AudioEnabled         bool   `json:"audioEnabled"`
		BackupType           string `json:"backupType"`
		ControlEnabled       bool   `json:"controlEnabled"`
		DewarpingParams      string `json:"dewarpingParams"`
		DisableDualStreaming bool   `json:"disableDualStreaming"`
		FailoverPriority     string `json:"failoverPriority"`
		GroupID              string `json:"groupId"`
		GroupName            string `json:"groupName"`
		ID                   string `json:"id"`
		LicenseUsed          bool   `json:"licenseUsed"`
		LogicalID            string `json:"logicalId"`
		Mac                  string `json:"mac"`
		ManuallyAdded        bool   `json:"manuallyAdded"`
		MaxArchiveDays       int    `json:"maxArchiveDays"`
		MinArchiveDays       int    `json:"minArchiveDays"`
		Model                string `json:"model"`
		MotionMask           string `json:"motionMask"`
		MotionType           string `json:"motionType"`
		Name                 string `json:"name"`
		ParentID             string `json:"parentId"`
		PhysicalID           string `json:"physicalId"`
		PreferredServerID    string `json:"preferredServerId"`
		ScheduleEnabled      bool   `json:"scheduleEnabled"`
		ScheduleTasks        []struct {
			AfterThreshold  int    `json:"afterThreshold"`
			BeforeThreshold int    `json:"beforeThreshold"`
			BitrateKbps     int    `json:"bitrateKbps"`
			DayOfWeek       int    `json:"dayOfWeek"`
			EndTime         int    `json:"endTime"`
			Fps             int    `json:"fps"`
			RecordAudio     bool   `json:"recordAudio"`
			RecordingType   string `json:"recordingType"`
			StartTime       int    `json:"startTime"`
			StreamQuality   string `json:"streamQuality"`
		} `json:"scheduleTasks"`
		Status               string `json:"status"`
		StatusFlags          string `json:"statusFlags"`
		TypeID               string `json:"typeId"`
		URL                  string `json:"url"`
		UserDefinedGroupName string `json:"userDefinedGroupName"`
		Vendor               string `json:"vendor"`
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	response := Response{}
	err = json.Unmarshal(b, &response)
	if err != nil {
		return nil, err
	}

	dest := make(map[string]string)
	for _, dev := range response {
		dest[dev.ID] = dev.Name
	}

	return dest, nil
}
