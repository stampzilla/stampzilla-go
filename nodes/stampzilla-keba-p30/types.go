package main

type Report2 struct {
	ID                  string `json:"ID"`
	State               int    `json:"State"`
	Error1              int    `json:"Error1"`
	Error2              int    `json:"Error2"`
	Plug                int    `json:"Plug"`
	AuthON              int    `json:"AuthON"`
	Authreq             int    `json:"Authreq"`
	EnableSys           int    `json:"Enable sys"`
	EnableUser          int    `json:"Enable user"`
	MaxCurr             int    `json:"Max curr"`
	MaxCurrPercent      int    `json:"Max curr %"`
	CurrHW              int    `json:"Curr HW"`
	CurrUser            int    `json:"Curr user"`
	CurrFS              int    `json:"Curr FS"`
	TmoFS               int    `json:"Tmo FS"`
	CurrTimer           int    `json:"Curr timer"`
	TmoCT               int    `json:"Tmo CT"`
	Setenergy           int    `json:"Setenergy"`
	Output              int    `json:"Output"`
	Input               int    `json:"Input"`
	X2PhaseSwitchSource int    `json:"X2 phaseSwitch source"`
	X2PhaseSwitch       int    `json:"X2 phaseSwitch"`
	Serial              string `json:"Serial"`
	Sec                 int    `json:"Sec"`
}
type Report3 struct {
	ID     string `json:"ID"`
	U1     int    `json:"U1"`
	U2     int    `json:"U2"`
	U3     int    `json:"U3"`
	I1     int    `json:"I1"`
	I2     int    `json:"I2"`
	I3     int    `json:"I3"`
	P      int    `json:"P"`
	PF     int    `json:"PF"`
	EPres  int    `json:"E pres"`
	ETotal int    `json:"E total"`
	Serial string `json:"Serial"`
	Sec    int    `json:"Sec"`
}
