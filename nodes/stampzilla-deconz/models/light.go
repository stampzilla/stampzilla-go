package models

// Lights is a list of Light's
type Lights map[string]Light

func NewLights() Lights {
	return make(map[string]Light)
}

// Light is a deconz RESP api light
type Light struct {
	Etag         string `json:"etag"`
	Hascolor     bool   `json:"hascolor"`
	Manufacturer string `json:"manufacturer"`
	Modelid      string `json:"modelid"`
	Name         string `json:"name"`
	State        map[string]interface{}
	//Pointsymbol  struct {
	//} `json:"pointsymbol"`
	//State struct {
	//Alert     string    `json:"alert"`
	//Bri       int       `json:"bri"`
	//Colormode string    `json:"colormode"`
	//Ct        int       `json:"ct"`
	//Effect    string    `json:"effect"`
	//Hue       int       `json:"hue"`
	//On        bool      `json:"on"`
	//Reachable bool      `json:"reachable"`
	//Sat       int       `json:"sat"`
	//Xy        []float64 `json:"xy"`
	//} `json:"state"`
	Swversion string `json:"swversion"`
	Type      string `json:"type"`
	Uniqueid  string `json:"uniqueid"`
}
