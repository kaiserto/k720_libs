package main

import (
	"fmt"
	"k720"

	"go.bug.st/serial"
)

var macAddr byte = 0x0f
var port string = "/dev/ttyS0"

var comHandle serial.Port

func Test1() {
	// Reset
	// k720.SendCmd(comHandle, macAddr, "RS")

	// Card Read Position
	k720.SendCmd(comHandle, macAddr, "FC7")

	/////////////////////////////////////////////////////////////////////////

	version, err := k720.GetSysVersion(comHandle, macAddr)
	if err == nil {
		fmt.Println(string(version))
	}

	query, err := k720.Query(comHandle, macAddr)
	if err == nil {
		k720.PrintPacket("", query)
	}

	sensorQuery, err := k720.SensorQuery(comHandle, macAddr)
	if err == nil {
		k720.PrintPacket("", sensorQuery)
	}

	/////////////////////////////////////////////////////////////////////////

	fmt.Println("== S50 ==")

	data, err := k720.S50DetectCard(comHandle, macAddr)
	if err == nil {
		fmt.Println(string(data))
	}

	data, err = k720.S50GetCardId(comHandle, macAddr)
	if (err == nil) && (data[0] == byte('P')) {
		cardId := data[3:]
		k720.PrintPacket("Card ID: ", cardId)
	}

	data, err = k720.S50LoadSecKey(comHandle, macAddr, 0x00, k720.KEYA, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	if (err == nil) && (data[0] == byte('P')) {
		fmt.Println("Password check successfull")
	} else {
		fmt.Println("Password check failure")
	}

	data, err = k720.S50LoadSecKey(comHandle, macAddr, 0x00, k720.KEYA, []byte{0x00, 0xff, 0xff, 0xff, 0xff, 0xff})
	if (err == nil) && (data[0] == byte('P')) {
		fmt.Println("Password check successfull")
	} else {
		fmt.Println("Password check failure")
	}

	/////////////////////////////////////////////////////////////////////////

	fmt.Println("== S70 ==")

	data, err = k720.S70DetectCard(comHandle, macAddr)
	if err == nil {
		fmt.Println(string(data))
	}

	data, err = k720.S70GetCardId(comHandle, macAddr)
	if (err == nil) && (data[0] == byte('P')) {
		cardId := data[3:]
		k720.PrintPacket("Card ID: ", cardId)
	}

	/////////////////////////////////////////////////////////////////////////

	fmt.Println("== UL ==")

	data, err = k720.ULDetectCard(comHandle, macAddr)
	if err == nil {
		fmt.Println(string(data))
	}

	data, err = k720.ULGetCardId(comHandle, macAddr)
	if (err == nil) && (data[0] == byte('P')) {
		cardId := data[3:]
		k720.PrintPacket("Card ID: ", cardId)
	}
}

func CardPositions() {
	// Send Card
	// k720.SendCmd(comHandle, macAddr, "DC")
	// Recycle
	// k720.SendCmd(comHandle, macAddr, "CP")

	// Outside Position
	// k720.SendCmd(comHandle, macAddr, "FC0")
	// Take Card Position
	// k720.SendCmd(comHandle, macAddr, "FC4")
	// Sensor 2 Position
	// k720.SendCmd(comHandle, macAddr, "FC6")
	// Card Read Position
	// k720.SendCmd(comHandle, macAddr, "FC7")
	// FrontEnterCard
	k720.SendCmd(comHandle, macAddr, "FC8")

	sensorQuery, err := k720.SensorQuery(comHandle, macAddr)
	if err == nil {
		k720.PrintPacket("", sensorQuery)
	}
}

func Operate() {
	// Reset
	k720.SendCmd(comHandle, macAddr, "RS")
	for {
		// Outside Position
		k720.SendCmd(comHandle, macAddr, "FC0")
		for {
			sensorQuery, _ := k720.SensorQuery(comHandle, macAddr)
			state := k720.CalculateState(sensorQuery)
			if state&k720.CARD_AT_SENSOR_1_POSITION > 0 {
				break
			}
			// FrontEnterCard
			k720.SendCmd(comHandle, macAddr, "FC8")
			// Card Read Position
			k720.SendCmd(comHandle, macAddr, "FC7")
		}
		data, err := k720.S50GetCardId(comHandle, macAddr)
		if (err == nil) && (data[0] == byte('P')) {
			cardId := data[3:]
			k720.PrintPacket("Card ID: ", cardId)
		}

		//  Take Card Position
		k720.SendCmd(comHandle, macAddr, "FC4")
	}
}

func main() {
	k720.LOG_LEVEL = k720.LOG_NONE

	comHandle, _ = k720.CommOpen(port)
	// Test1()
	// CardPositions()
	Operate()
	k720.CommClose(comHandle)
}
