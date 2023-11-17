import serial

STX = 0x02
ETX = 0x03
ENQ = 0x05
ACK = 0x06
NAK = 0x15

LOG_NONE, LOG_TRACE, LOG_DEBUG, LOG_INFO, LOG_WARN, LOG_ERROR, LOG_FATAL = \
    (0, 1, 2, 3, 4, 5, 6)
LOG_LEVEL = LOG_NONE

CARD_AT_SENSOR_1_POSITION = 0x0001
CARD_AT_SENSOR_2_POSITION = 0x0002
CARD_AT_SENSOR_3_POSITION = 0x0004

KEYA = 0x30
KEYB = 0x31

def _log_trace(*objects):
    if LOG_LEVEL == LOG_NONE:
        return
    if LOG_LEVEL <= LOG_TRACE:
        print(*objects)

def _log_debug(*objects):
    if LOG_LEVEL == LOG_NONE:
        return
    if LOG_LEVEL <= LOG_DEBUG:
        print(*objects)

def _debug_packet(prefix, packet):
    if LOG_LEVEL == LOG_NONE:
        return
    if LOG_LEVEL <= LOG_DEBUG:
        print_packet(prefix, packet)

def _calculate_bcc(packet):
    # type: (list) -> int
    xorsum = 0x00
    for b in packet:
        xorsum ^= b
    return xorsum

def _create_packet(mac_addr, data):
    # type: (int, str) -> list | None
    packet = [] 
    selen = len(data)
    if mac_addr < 0 or mac_addr > 15:
        return None
    if selen == 0:
        return None

    # start
    packet.append(STX)

    # address
    addr_h = ord('0') + mac_addr // 10
    addr_l = ord('0') + mac_addr % 10
    packet.append(addr_h)
    packet.append(addr_l)

    # selen
    selen_l = selen & 0xff
    selen_h = (selen << 8) & 0xff
    packet.append(selen_h)
    packet.append(selen_l)
    # data
    for d in data:
        packet.append(ord(d))

    # end
    packet.append(ETX)

    # bcc
    bcc = _calculate_bcc(packet)
    packet.append(bcc)

    return packet

def _create_enquiry_packet(mac_addr):
    # type: (int) -> list | None 
    packet = []
    if mac_addr < 0 or mac_addr > 15:
        return None
    # enquiry
    packet.append(ENQ)

    # address
    addr_h = ord('0') + (mac_addr // 10)
    addr_l = ord('0') + (mac_addr % 10)
    packet.append(addr_h)
    packet.append(addr_l)

    return packet

def _read_ack(com_handle, mac_addr):
    # type: (object, int) -> bool
    b_ack = com_handle.read(size=1)
    if len(b_ack) != 1:
        return False
    ack = b_ack[0]    
    if ack != ACK:
        return False

    b_addr = com_handle.read(size=2)
    _log_trace("Received address:", b_addr.decode())
    if b_addr.decode() != str(mac_addr):
        return False
    return True

def _send_packet(com_handle, mac_addr, command):
    # type: (object, int, str) -> list | None
    packet = _create_packet(mac_addr, command)
    if packet == None:
        return None
    _debug_packet('sent: ', packet)
    com_handle.write(packet)

    if not _read_ack(com_handle, mac_addr):
        return None

    packet = _create_enquiry_packet(mac_addr)
    if packet == None:
        return None
    com_handle.write(packet)

    packet = _receive_packet(com_handle, mac_addr)
    if packet == None:
        return None
    return packet

def _receive_packet(com_handle, mac_addr):
    # type: (object, int) -> list | None
    packet = []
    # start
    b_stx = com_handle.read(size=1)
    _log_trace("bSTX:" + str(bytes(b_stx)))
    if len(b_stx) != 1:
        return None
    stx = b_stx[0]
    if stx != STX:
        return None
    packet.append(stx)

    # address
    b_addr = com_handle.read(size=2)
    _log_trace("bADDR:" + str(bytes(b_addr)))
    if len(b_addr) != 2:
        return None
    if b_addr.decode() != str(mac_addr):
        return None
    packet.append(b_addr[0])
    packet.append(b_addr[1])

    # relen
    b_relen = com_handle.read(size=2)
    _log_trace("bRELEN:" + str(bytes(b_relen)))
    if len(b_relen) != 2:
        return None
    relen_h = b_relen[0]
    relen_l = b_relen[1]
    relen = 0xff * relen_h + relen_l
    packet.append(relen_h)
    packet.append(relen_l)
    # data
    b_data = com_handle.read(size=relen)
    _log_trace("bDATA:" + str(bytes(b_data)))
    if len(b_data) != relen:
        return None
    for d in b_data:
        packet.append(d)

    # end
    b_etx = com_handle.read(size=1)
    _log_trace("bETX:" + str(bytes(b_etx)))
    if len(b_etx) != 1:
        return None
    etx = b_etx[0]
    if etx != ETX:
        return None
    packet.append(etx)
    
    # bcc
    b_bcc = com_handle.read(size=1)
    _log_trace("bBCC:" + str(bytes(b_bcc)))
    if len(b_bcc) != 1:
        return None
    bcc = b_bcc[0]
    if bcc != _calculate_bcc(packet):
        return None
    packet.append(bcc)

    _debug_packet('received: ', packet)
    return packet

def _get_packet_data(packet):
    # type: (list) -> list | None
    data = []
    if len(packet) < 5:
        return None

    # relen
    relen_h = packet[3]
    relen_l = packet[4]
    relen = 0xff * relen_h + relen_l
    # data
    data = packet[5:5+relen]
    return data

#############################################################################

def print_packet(prefix, packet):
    # type: (str, list) -> None
    hex_sequence = ' '.join(['0x' + format(b, '02x') for b in packet])
    print(prefix + hex_sequence)

def calculate_state(query):
    # type: (list) -> int
    state = 0x0000
    for b in query:
        s = b - ord('0')
        state <<= 4
        state |= s
    # print(hex(state))
    return state

def comm_open(port):
    # type: (str) -> object
    return serial.Serial(
        port=port,
        baudrate=9600,
        bytesize=serial.EIGHTBITS,
        parity=serial.PARITY_NONE,
        stopbits=serial.STOPBITS_ONE,
        timeout=1)

def comm_open_with_baud(port, baudrate):
    # type: (str, int) -> object
    return serial.Serial(
        port=port,
        baudrate=baudrate,
        bytesize=serial.EIGHTBITS,
        parity=serial.PARITY_NONE,
        stopbits=serial.STOPBITS_ONE,
        timeout=1)

def comm_close(com_handle):
    # type: (object) -> None
    com_handle.close()

# Get Version (GV)
def get_sys_version(com_handle, mac_addr):
    # type: (object, int) -> str | None
    packet = _send_packet(com_handle, mac_addr, 'GV')
    if packet == None:
        return None
    data = _get_packet_data(packet)
    if data == None:
        return None
    return bytes(data).decode()

# Check Sensor (RF)
def query(com_handle, mac_addr):
    # type: (object, int) -> list | None
    packet = _send_packet(com_handle, mac_addr, 'RF')
    if packet == None:
        return None
    query = _get_packet_data(packet)
    if len(query) < 3:
        return None
    return query[2:]

# Advanced Check (AP)
def sensor_query(com_handle, mac_addr):
    # type: (object, int) -> list | None
    packet = _send_packet(com_handle, mac_addr, 'AP')
    if packet == None:
        return None
    sensor_query = _get_packet_data(packet)
    if len(sensor_query) < 3:
        return None
    return sensor_query[2:]

def send_cmd(com_handle, mac_addr, command):
    # type: (object, int, str) -> list | None
    packet = _send_packet(com_handle, mac_addr, command)
    if packet == None:
        return None
    return _get_packet_data(packet)

def s50_detect_card(com_handle, mac_addr):
    # type: (object, int) -> list | None
    packet = _send_packet(com_handle, mac_addr, '\x3b\x30')
    if packet == None:
        return None
    return _get_packet_data(packet)

def s50_get_card_id(com_handle, mac_addr):
    # type: (object, int) -> list | None
    packet = _send_packet(com_handle, mac_addr, '\x3b\x31')
    if packet == None:
        return None
    return _get_packet_data(packet)

def s50_load_sec_key(com_handle, mac_addr, sector_addr, key_type, key):
    # type: (object, int, int, int, list) -> list | None
    command = '\x3b\x32' + \
            chr(sector_addr) + \
            chr(key_type) + \
            ''.join([chr(b) for b in key])
    packet = _send_packet(com_handle, mac_addr, command)
    if packet == None:
        return None
    return _get_packet_data(packet)

def s70_detect_card(com_handle, mac_addr):
    # type: (object, int) -> list | None
    packet = _send_packet(com_handle, mac_addr, '\x3c\x30')
    if packet == None:
        return None
    return _get_packet_data(packet)

def s70_get_card_id(com_handle, mac_addr):
    # type: (object, int) -> list | None
    packet = _send_packet(com_handle, mac_addr, '\x3c\x31')
    if packet == None:
        return None
    return _get_packet_data(packet)

def ul_detect_card(com_handle, mac_addr):
    # type: (object, int) -> list | None
    packet = _send_packet(com_handle, mac_addr, '\x3d\x30')
    if packet == None:
        return None
    return _get_packet_data(packet)

def ul_get_card_id(com_handle, mac_addr):
    # type: (object, int) -> list | None
    packet = _send_packet(com_handle, mac_addr, '\x3d\x31')
    if packet == None:
        return None
    return _get_packet_data(packet)
