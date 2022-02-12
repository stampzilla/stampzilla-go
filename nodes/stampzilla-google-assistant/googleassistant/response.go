package googleassistant

type DeviceName struct {
	DefaultNames []string `json:"defaultNames,omitempty"`
	Name         string   `json:"name,omitempty"`
	Nicknames    []string `json:"nicknames,omitempty"`
}

type DeviceAttributes struct {
	ColorModel      string `json:"colorModel,omitempty"`
	TemperatureMinK int    `json:"temperatureMinK,omitempty"`
	TemperatureMaxK int    `json:"temperatureMaxK,omitempty"`
}

type Device struct {
	ID              string     `json:"id"`
	Type            string     `json:"type"`
	Traits          []string   `json:"traits"`
	Name            DeviceName `json:"name"`
	WillReportState bool       `json:"willReportState"`
	// DeviceInfo      struct {
	// Manufacturer string `json:"manufacturer"`
	// Model        string `json:"model"`
	// HwVersion    string `json:"hwVersion"`
	// SwVersion    string `json:"swVersion"`
	// } `json:"deviceInfo"`
	// CustomData struct {
	// FooValue int    `json:"fooValue"`
	// BarValue bool   `json:"barValue"`
	// BazValue string `json:"bazValue"`
	// } `json:"customData"`
	Attributes DeviceAttributes `json:"attributes,omitempty"`
}

type ResponseStates struct {
	On         bool `json:"on,omitempty"`
	Brightness int  `json:"brightness,omitempty"`
	Online     bool `json:"online,omitempty"`
}

func NewResponseCommand() ResponseCommand {
	return ResponseCommand{
		States: ResponseStates{},
		Status: "SUCCESS",
	}
}

type ResponseCommand struct {
	IDs       []string       `json:"ids"`
	Status    string         `json:"status"`
	States    ResponseStates `json:"states"`
	ErrorCode string         `json:"errorCode,omitempty"`
}

type Response struct {
	RequestID string `json:"requestId"`
	Payload   struct {
		AgentUserID string            `json:"agentUserId,omitempty"`
		Devices     []Device          `json:"devices,omitempty"`
		Commands    []ResponseCommand `json:"commands,omitempty"`
	} `json:"payload"`
}

type QueryResponse struct {
	RequestID string `json:"requestId"`
	Payload   struct {
		Devices map[string]map[string]interface{} `json:"devices,omitempty"`
	} `json:"payload"`
}
