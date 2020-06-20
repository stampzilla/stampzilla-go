package main

import (
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
)

/*
	No Unit / Data type Additional info
	0 Degrees Divide by 10
	1 On/off bool 0 or 1
	2 Number Divide by 10
	3 Percent Divide by 10
	4 Ampere Divide by 10
	5 kWh Divide by 10
	6 Hours As is
	7 Minutes As is
	8 Degree minutes As is
	9 kw As is
	A Pulses As is (For S0 El-meter pulse counter)
*/

type HeatPump struct {
	RoomSensorInfluence   int `json:"2204"` // Number /10
	AddHeatStatus         int `json:"3104"`
	ExtraWarmWater        int `json:"6209"` // Hours
	RadiatorForward       int `json:"0002"` // Degrees /10
	HeatCarrierReturn     int `json:"0003"` // Degrees /10
	HeatCarrierForward    int `json:"0004"` // Degrees /10
	BrineIn               int `json:"0005"` // Degrees /10
	BrineOut              int `json:"0006"` // Degrees /10
	Outdoor               int `json:"0007"` // Degrees /10
	Indoor                int `json:"0008"` // Degrees /10
	WarmWater1Top         int `json:"0009"` // Degrees /10
	HotGasCompressor      int `json:"000B"` // Degrees /10
	AirIntake             int `json:"000E"` // Degrees /10
	Pool                  int `json:"0011"` // Degrees /10
	RadiatorForward2      int `json:"0020"` // Degrees /10
	Indoor2               int `json:"0021"` // Degrees /10
	Compressor            int `json:"1A01"` // Bool
	PumpColdCircuit       int `json:"1A04"` // Bool
	PumpHeatCircuit       int `json:"1A05"` // Bool
	PumpRadiator          int `json:"1A06"` // Bool
	SwitchValve1          int `json:"1A07"` // Bool
	SwitchValve2          int `json:"1A08"` // Bool
	Fan                   int `json:"1A09"` // Bool
	HighPressostat        int `json:"1A0A"` // Bool
	LowPressostat         int `json:"1A0B"` // Bool
	HeatingCable          int `json:"1A0C"` // Bool
	CrankCaseHeater       int `json:"1A0D"` // Bool
	Alarm                 int `json:"1A20"` // Bool
	PumpRadiator2         int `json:"1A21"` // Bool
	WarmWaterSetpoint     int `json:"0111"` // Degrees /10
	HeatingSetpoint       int `json:"0107"` // Degrees /10
	HeatingSetpoint2      int `json:"0120"` // Degrees /10
	RoomTempSetpoint      int `json:"0203"` // Degrees /10
	HeatSet1CurveL        int `json:"0205"` // Degrees /10
	HeatSet2CurveR        int `json:"0206"` // Degrees /10
	HeatSet1CurveL2       int `json:"0222"` // Degrees /10
	HeatSet2CurveR2       int `json:"0223"` // Degrees /10
	PoolTempSetpoint      int `json:"0219"` // Degrees /10
	CollectedPulsesMeter1 int `json:"AFF1"`
	CollectedPulsesMeter2 int `json:"AFF2"`
}

func (hp HeatPump) State() devices.State {
	state := make(devices.State)

	state["RoomSensorInfluence"] = float64(hp.RoomSensorInfluence) / 10.0
	state["AddHeatStatus"] = hp.AddHeatStatus
	state["ExtraWarmWater"] = hp.ExtraWarmWater
	state["RadiatorForward"] = float64(hp.RadiatorForward) / 10.0
	state["HeatCarrierReturn"] = float64(hp.HeatCarrierReturn) / 10.0
	state["HeatCarrierForward"] = float64(hp.HeatCarrierForward) / 10.0
	state["BrineIn"] = float64(hp.BrineIn) / 10.0
	state["BrineOut"] = float64(hp.BrineOut) / 10.0
	state["Outdoor"] = float64(hp.Outdoor) / 10.0
	state["Indoor"] = float64(hp.Indoor) / 10.0
	state["WarmWater1Top"] = float64(hp.WarmWater1Top) / 10.0
	state["HotGasCompressor"] = float64(hp.HotGasCompressor) / 10.0
	state["AirIntake"] = float64(hp.AirIntake) / 10.0
	state["Pool"] = float64(hp.Pool) / 10.0
	state["RadiatorForward2"] = float64(hp.RadiatorForward2) / 10.0
	state["Indoor2"] = float64(hp.Indoor2) / 10.0
	state["Compressor"] = hp.Compressor != 0
	state["PumpColdCircuit"] = hp.PumpColdCircuit != 0
	state["PumpHeatCircuit"] = hp.PumpHeatCircuit != 0
	state["PumpRadiator"] = hp.PumpRadiator != 0
	state["SwitchValve1"] = hp.SwitchValve1 != 0
	state["SwitchValve2"] = hp.SwitchValve2 != 0
	state["Fan"] = hp.Fan != 0
	state["HighPressostat"] = hp.HighPressostat != 0
	state["LowPressostat"] = hp.LowPressostat != 0
	state["HeatingCable"] = hp.HeatingCable != 0
	state["CrankCaseHeater"] = hp.CrankCaseHeater != 0
	state["Alarm"] = hp.Alarm != 0
	state["PumpRadiator2"] = hp.PumpRadiator2 != 0
	state["WarmWaterSetpoint"] = float64(hp.WarmWaterSetpoint) / 10.0
	state["HeatingSetpoint"] = float64(hp.HeatingSetpoint) / 10.0
	state["HeatingSetpoint2"] = float64(hp.HeatingSetpoint2) / 10.0
	state["RoomTempSetpoint"] = float64(hp.RoomTempSetpoint) / 10.0
	state["HeatSet1CurveL"] = float64(hp.HeatSet1CurveL) / 10.0
	state["HeatSet2CurveR"] = float64(hp.HeatSet2CurveR) / 10.0
	state["HeatSet1CurveL2"] = float64(hp.HeatSet1CurveL2) / 10.0
	state["HeatSet2CurveR2"] = float64(hp.HeatSet2CurveR2) / 10.0
	state["PoolTempSetpoint"] = float64(hp.PoolTempSetpoint) / 10.0
	state["CollectedPulsesMeter1"] = hp.CollectedPulsesMeter1
	state["CollectedPulsesMeter2"] = hp.CollectedPulsesMeter2

	return state
}
