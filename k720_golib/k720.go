package k720

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.bug.st/serial"
)

const (
	STX byte = 0x02
	ETX byte = 0x03
	ENQ byte = 0x05
	ACK byte = 0x06
	NAK byte = 0x15
)

const (
	LOG_NONE  int = 0 // iota
	LOG_TRACE int = 1 // iota
	LOG_DEBUG int = 2 // iota
	LOG_INFO  int = 3 // iota
	LOG_WARN  int = 4 // iota
	LOG_ERROR int = 5 // iota
	LOG_FATAL int = 6 // iota
)

var LOG_LEVEL int = LOG_NONE

type State struct {
	Code        int
	Description string
}

var (
	CARD_AT_SENSOR_1_POSITION  State = registerState(0x0001, "Card at sensor 1 position")
	CARD_AT_SENSOR_2_POSITION  State = registerState(0x0002, "Card at sensor 2 position")
	CARD_AT_SENSOR_3_POSITION  State = registerState(0x0004, "Card at sensor 3 position")
	CARD_EMPTY                 State = registerState(0x0008, "Card empty")
	CARD_PRE_EMPTY             State = registerState(0x0010, "Card pre-empty")
	CARD_JAM                   State = registerState(0x0020, "Card jam")
	CARD_OVERLAP               State = registerState(0x0040, "Card overlap")
	CARD_HOPPER_FULL           State = registerState(0x0080, "Card hopper full")
	ERROR_OF_RECYCLING_CARD    State = registerState(0x0100, "Error of recycling card")
	ERROR_OF_ISSUING_CARD      State = registerState(0x0200, "Error of issuing card")
	COLLECTING_CARD            State = registerState(0x0400, "Collecting card")
	SENDING_CARD               State = registerState(0x0800, "Sending card")
	PREPARING_CARD             State = registerState(0x1000, "Preparing card")
	PREPARE_CARD_FAILURE       State = registerState(0x2000, "Prepare card failure")
	COULD_NOT_IMPLEMNT_COMMAND State = registerState(0x4000, "Could not implement command")
	RECYCLING_BOX_FULL         State = registerState(0x8000, "Recycling box fulls")
)
var states []State = make([]State, 0)

func registerState(code int, description string) State {
	state := State{code, description}
	states = append(states, state)
	return state
}

const (
	KEYA byte = 0x30
	KEYB byte = 0x31
)

func logTrace(args ...interface{}) {
	if LOG_LEVEL == LOG_NONE {
		return
	}
	if LOG_LEVEL <= LOG_TRACE {
		fmt.Println(args...)
	}
}

func logDebug(args ...interface{}) {
	if LOG_LEVEL == LOG_NONE {
		return
	}
	if LOG_LEVEL <= LOG_DEBUG {
		fmt.Println(args...)
	}
}

func debugPacket(prefix string, packet []byte) {
	if LOG_LEVEL == LOG_NONE {
		return
	}
	if LOG_LEVEL <= LOG_DEBUG {
		PrintPacket(prefix, packet)
	}
}

func calculateBcc(packet []byte) byte {
	xorsum := byte(0x00)
	for _, b := range packet {
		xorsum ^= b
	}
	return xorsum
}

func hexEncodeBytes(bytes []byte) string {
	str := ""
	for _, b := range bytes {
		str += fmt.Sprintf("0x%02x ", b)
	}
	return str
}

func commRead(comHandle serial.Port, p []byte) (n int, err error) {
	i := 0
	buf := make([]byte, 1)
	for {
		buf[0] = 0x00
		n, err := comHandle.Read(buf)
		if err != nil {
			return i, err
		}
		if n == 0 {
			return i, nil
		}
		// copy(p[i:], buf[:])
		copy(p[i:], buf)
		i += n
		if i == len(p) {
			return i, nil
		}
	}
}

func createPacket(macAddr byte, data string) ([]byte, error) {
	packet := make([]byte, 0)
	selen := len(data)
	if /*macAddr < 0 ||*/ macAddr > 15 {
		return nil, errors.New("error #1 in createPacket")
	}
	if selen == 0 {
		return nil, errors.New("error #2 in createPacket")
	}

	// start
	packet = append(packet, STX)

	// address
	addrH := int('0') + int(macAddr/10)
	addrL := int('0') + int(macAddr%10)
	packet = append(packet, byte(addrH))
	packet = append(packet, byte(addrL))

	// selen
	selenL := selen & 0xff
	selenH := (selen << 8) & 0xff
	packet = append(packet, byte(selenH))
	packet = append(packet, byte(selenL))
	// data
	packet = append(packet, data...)

	// end
	packet = append(packet, ETX)

	// bcc
	bcc := calculateBcc(packet)
	packet = append(packet, bcc)

	return packet, nil
}

func createEnquiryPacket(macAddr byte) ([]byte, error) {
	packet := make([]byte, 0)
	if /*macAddr < 0 ||*/ macAddr > 15 {
		return nil, errors.New("error #1 in createEnquiryPacket")
	}
	// enquiry
	packet = append(packet, ENQ)

	// address
	addrH := int('0') + int(macAddr/10)
	addrL := int('0') + int(macAddr%10)
	packet = append(packet, byte(addrH))
	packet = append(packet, byte(addrL))

	return packet, nil
}

func readAck(comHandle serial.Port, macAddr byte) error {
	arrAck := make([]byte, 1)
	n, err := commRead(comHandle, arrAck)
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("error #1 in readAck")
	}
	ack := arrAck[0]
	if ack != ACK {
		return errors.New("error #2 in readAck")
	}

	arrAddr := make([]byte, 2)
	n, err = commRead(comHandle, arrAddr)
	if err != nil {
		return err
	}
	if n != 2 {
		return errors.New("error #3 in readAck")
	}
	logTrace("Received address:", string(arrAddr))
	if string(arrAddr) != strconv.Itoa(int(macAddr)) {
		return errors.New("error #4 in readAck")
	}
	return nil
}

func sendPacket(comHandle serial.Port, macAddr byte, command string) ([]byte, error) {
	packet, err := createPacket(macAddr, command)
	debugPacket("sent: ", packet)
	if err != nil {
		return nil, err
	}
	comHandle.Write(packet)

	err = readAck(comHandle, macAddr)
	if err != nil {
		return nil, err
	}

	packet, err = createEnquiryPacket(macAddr)
	if err != nil {
		return nil, err
	}
	comHandle.Write(packet)

	packet, err = receivePacket(comHandle, macAddr)
	if err != nil {
		return nil, err
	}
	return packet, nil
}

func receivePacket(comHandle serial.Port, macAddr byte) ([]byte, error) {
	packet := make([]byte, 0)
	// start
	arrStx := make([]byte, 1)
	// n, err := commRead(comHandle, arrStx)
	n, err := comHandle.Read(arrStx)
	logTrace("bSTX: " + hexEncodeBytes(arrStx))
	if err != nil {
		return nil, err
	}
	if n != 1 {
		return nil, errors.New("error #1 in receivePacket")
	}
	stx := arrStx[0]
	if stx != STX {
		return nil, errors.New("error #2 in receivePacket")
	}
	packet = append(packet, stx)

	// address
	arrAddr := make([]byte, 2)
	n, err = commRead(comHandle, arrAddr)
	logTrace("bADDR: " + hexEncodeBytes(arrAddr))
	if err != nil {
		return nil, err
	}
	if n != 2 {
		return nil, errors.New("error #3 in receivePacket")
	}
	if string(arrAddr) != strconv.Itoa(int(macAddr)) {
		return nil, errors.New("error #4 in receivePacket")
	}
	packet = append(packet, arrAddr[0])
	packet = append(packet, arrAddr[1])

	// relen
	arrRelen := make([]byte, 2)
	n, err = commRead(comHandle, arrRelen)
	logTrace("bRELEN: " + hexEncodeBytes(arrRelen))
	if err != nil {
		return nil, err
	}
	if n != 2 {
		return nil, errors.New("error #5 in receivePacket")
	}
	relenH := arrRelen[0]
	relenL := arrRelen[1]
	relen := 0xff*relenH + relenL
	packet = append(packet, relenH)
	packet = append(packet, relenL)
	// data
	arrData := make([]byte, relen)
	n, err = commRead(comHandle, arrData)
	logTrace("bDATA: " + hexEncodeBytes(arrData))
	if err != nil {
		return nil, err
	}
	if n != int(relen) {
		return nil, errors.New("error #6 in receivePacket")
	}
	packet = append(packet, arrData...)

	// end
	arrEtx := make([]byte, 1)
	n, err = commRead(comHandle, arrEtx)
	logTrace("bETX: " + hexEncodeBytes(arrEtx))
	if err != nil {
		return nil, err
	}
	if n != 1 {
		return nil, errors.New("error #7 in receivePacket")
	}
	etx := arrEtx[0]
	if etx != ETX {
		return nil, errors.New("error #8 in receivePacket")
	}
	packet = append(packet, etx)

	// bcc
	bccEtx := make([]byte, 1)
	n, err = commRead(comHandle, bccEtx)
	logTrace("bBCC: " + hexEncodeBytes(bccEtx))
	if err != nil {
		return nil, err
	}
	if n != 1 {
		return nil, errors.New("error #9 in receivePacket")
	}
	bcc := bccEtx[0]
	if bcc != calculateBcc(packet) {
		return nil, errors.New("error #10 in receivePacket")
	}
	packet = append(packet, bcc)

	debugPacket("received: ", packet)
	return packet, nil
}

func getPacketData(packet []byte) ([]byte, error) {
	if len(packet) < 5 {
		return nil, errors.New("error #1 in getPacketData")
	}

	// relen
	relenH := packet[3]
	relenL := packet[4]
	relen := 0xff*relenH + relenL
	// data
	data := packet[5 : 5+relen]
	return data, nil
}

/////////////////////////////////////////////////////////////////////////////

func PrintPacket(prefix string, packet []byte) {
	fmt.Println(prefix + hexEncodeBytes(packet))
}

func CalculateState(query []byte) int {
	state := 0x0000
	for _, b := range query {
		s := b - byte('0')
		state <<= 4
		state |= int(s)
	}
	return state
}

func PrintState(state int) {
	for _, s := range states {
		if state&s.Code > 0 {
			fmt.Println(s.Description)
		}
	}
}

func CommOpen(port string) (serial.Port, error) {
	mode := &serial.Mode{
		BaudRate: 9600,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
	comHandle, err := serial.Open(port, mode)
	comHandle.SetReadTimeout(1 * time.Second)
	return comHandle, err
}

func CommOpenWitBaud(port string, baudrate int) (serial.Port, error) {
	mode := &serial.Mode{
		BaudRate: baudrate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
	comHandle, err := serial.Open(port, mode)
	comHandle.SetReadTimeout(1 * time.Second)
	return comHandle, err
}

func CommClose(comHandle serial.Port) error {
	return comHandle.Close()
}

func GetSysVersion(comHandle serial.Port, macAddr byte) (string, error) {
	packet, err := sendPacket(comHandle, macAddr, "GV")
	if err != nil {
		return "", err
	}
	data, err := getPacketData(packet)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func Query(comHandle serial.Port, macAddr byte) ([]byte, error) {
	packet, err := sendPacket(comHandle, macAddr, "RF")
	if err != nil {
		return nil, err
	}
	query, err := getPacketData(packet)
	if err != nil {
		return nil, err
	}
	if len(query) < 3 {
		return nil, errors.New("error #1 in Query")
	}
	return query[2:], nil
}

func SensorQuery(comHandle serial.Port, macAddr byte) ([]byte, error) {
	packet, err := sendPacket(comHandle, macAddr, "AP")
	if err != nil {
		return nil, err
	}
	query, err := getPacketData(packet)
	if err != nil {
		return nil, err
	}
	if len(query) < 3 {
		return nil, errors.New("error #1 in SensorQuery")
	}
	return query[2:], nil
}

func SendCmd(comHandle serial.Port, macAddr byte, command string) ([]byte, error) {
	packet, err := sendPacket(comHandle, macAddr, command)
	if err != nil {
		return nil, err
	}
	return getPacketData(packet)
}

func S50DetectCard(comHandle serial.Port, macAddr byte) ([]byte, error) {
	packet, err := sendPacket(comHandle, macAddr, "\x3b\x30")
	if err != nil {
		return nil, err
	}
	return getPacketData(packet)
}

func S50GetCardId(comHandle serial.Port, macAddr byte) ([]byte, error) {
	packet, err := sendPacket(comHandle, macAddr, "\x3b\x31")
	if err != nil {
		return nil, err
	}
	return getPacketData(packet)
}

func S50LoadSecKey(comHandle serial.Port, macAddr byte, sectorAddr byte, keyType byte, key []byte) ([]byte, error) {
	command := "\x3b\x32" +
		string(sectorAddr) +
		string(keyType) +
		string(key)
	packet, err := sendPacket(comHandle, macAddr, command)
	if err != nil {
		return nil, err
	}
	return getPacketData(packet)
}

func S70DetectCard(comHandle serial.Port, macAddr byte) ([]byte, error) {
	packet, err := sendPacket(comHandle, macAddr, "\x3c\x30")
	if err != nil {
		return nil, err
	}
	return getPacketData(packet)
}

func S70GetCardId(comHandle serial.Port, macAddr byte) ([]byte, error) {
	packet, err := sendPacket(comHandle, macAddr, "\x3c\x31")
	if err != nil {
		return nil, err
	}
	return getPacketData(packet)
}

func ULDetectCard(comHandle serial.Port, macAddr byte) ([]byte, error) {
	packet, err := sendPacket(comHandle, macAddr, "\x3d\x30")
	if err != nil {
		return nil, err
	}
	return getPacketData(packet)
}

func ULGetCardId(comHandle serial.Port, macAddr byte) ([]byte, error) {
	packet, err := sendPacket(comHandle, macAddr, "\x3d\x31")
	if err != nil {
		return nil, err
	}
	return getPacketData(packet)
}
