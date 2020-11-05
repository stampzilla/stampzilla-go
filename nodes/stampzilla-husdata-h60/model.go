package main

import (
	"strconv"

	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
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

type number float64

func (ss *number) UnmarshalJSON(data []byte) error {
	i64, err := strconv.ParseUint(string(data), 10, 16)
	if err != nil {
		return err
	}
	*ss = number(float64(int16(i64)) / 10.0)
	return nil
}

type HeatPump struct {
	RoomSensorInfluence   number `json:"2204"` // Number /10
	AddHeatStatus         int    `json:"3104"`
	ExtraWarmWater        int    `json:"6209"` // Hours
	RadiatorForward       number `json:"0002"` // Degrees /10
	HeatCarrierReturn     number `json:"0003"` // Degrees /10
	HeatCarrierForward    number `json:"0004"` // Degrees /10
	BrineIn               number `json:"0005"` // Degrees /10
	BrineOut              number `json:"0006"` // Degrees /10
	Outdoor               number `json:"0007"` // Degrees /10
	Indoor                number `json:"0008"` // Degrees /10
	WarmWater1Top         number `json:"0009"` // Degrees /10
	HotGasCompressor      number `json:"000B"` // Degrees /10
	AirIntake             number `json:"000E"` // Degrees /10
	Pool                  number `json:"0011"` // Degrees /10
	RadiatorForward2      number `json:"0020"` // Degrees /10
	Indoor2               number `json:"0021"` // Degrees /10
	Compressor            int    `json:"1A01"` // Bool
	PumpColdCircuit       int    `json:"1A04"` // Bool
	PumpHeatCircuit       int    `json:"1A05"` // Bool
	PumpRadiator          int    `json:"1A06"` // Bool
	SwitchValve1          int    `json:"1A07"` // Bool
	SwitchValve2          int    `json:"1A08"` // Bool
	Fan                   int    `json:"1A09"` // Bool
	HighPressostat        int    `json:"1A0A"` // Bool
	LowPressostat         int    `json:"1A0B"` // Bool
	HeatingCable          int    `json:"1A0C"` // Bool
	CrankCaseHeater       int    `json:"1A0D"` // Bool
	Alarm                 int    `json:"1A20"` // Bool
	PumpRadiator2         int    `json:"1A21"` // Bool
	WarmWaterSetpoint     number `json:"0111"` // Degrees /10
	HeatingSetpoint       number `json:"0107"` // Degrees /10
	HeatingSetpoint2      number `json:"0120"` // Degrees /10
	RoomTempSetpoint      number `json:"0203"` // Degrees /10
	HeatSet1CurveL        number `json:"0205"` // Degrees /10
	HeatSet2CurveR        number `json:"0206"` // Degrees /10
	HeatSet1CurveL2       number `json:"0222"` // Degrees /10
	HeatSet2CurveR2       number `json:"0223"` // Degrees /10
	PoolTempSetpoint      number `json:"0219"` // Degrees /10
	CollectedPulsesMeter1 int    `json:"AFF1"`
	CollectedPulsesMeter2 int    `json:"AFF2"`
}

func (hp HeatPump) State() devices.State {
	state := make(devices.State)

	state["RoomSensorInfluence"] = hp.RoomSensorInfluence
	state["AddHeatStatus"] = hp.AddHeatStatus
	state["ExtraWarmWater"] = hp.ExtraWarmWater
	state["RadiatorForward"] = hp.RadiatorForward
	state["HeatCarrierReturn"] = hp.HeatCarrierReturn
	state["HeatCarrierForward"] = hp.HeatCarrierForward
	state["BrineIn"] = hp.BrineIn
	state["BrineOut"] = hp.BrineOut
	state["Outdoor"] = hp.Outdoor
	state["Indoor"] = hp.Indoor
	state["WarmWater1Top"] = hp.WarmWater1Top
	state["HotGasCompressor"] = hp.HotGasCompressor
	state["AirIntake"] = hp.AirIntake
	state["Pool"] = hp.Pool
	state["RadiatorForward2"] = hp.RadiatorForward2
	state["Indoor2"] = hp.Indoor2
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
	state["WarmWaterSetpoint"] = hp.WarmWaterSetpoint
	state["HeatingSetpoint"] = hp.HeatingSetpoint
	state["HeatingSetpoint2"] = hp.HeatingSetpoint2
	state["RoomTempSetpoint"] = hp.RoomTempSetpoint
	state["HeatSet1CurveL"] = hp.HeatSet1CurveL
	state["HeatSet2CurveR"] = hp.HeatSet2CurveR
	state["HeatSet1CurveL2"] = hp.HeatSet1CurveL2
	state["HeatSet2CurveR2"] = hp.HeatSet2CurveR2
	state["PoolTempSetpoint"] = hp.PoolTempSetpoint
	state["CollectedPulsesMeter1"] = hp.CollectedPulsesMeter1
	state["CollectedPulsesMeter2"] = hp.CollectedPulsesMeter2

	return state
}
