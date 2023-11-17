#! /usr/bin/env python3

from k720_pylib import k720

k720.LOG_LEVEL = k720.LOG_NONE

mac_addr = 0x0f
port = '/dev/ttyS0'

def test1():
    # Reset
    # k720.send_cmd(com_handle, mac_addr, 'RS')

    # Card Read Position
    k720.send_cmd(com_handle, mac_addr, 'FC7')

    #############################################################################

    version = k720.get_sys_version(com_handle, mac_addr)
    if version != None:
        print(version)

    query = k720.query(com_handle, mac_addr)
    if query != None:
        k720.print_packet('', query) 

    sensor_query = k720.sensor_query(com_handle, mac_addr)
    if sensor_query != None:
        k720.print_packet('', sensor_query)

    #############################################################################

    print('== S50 ==')

    data = k720.s50_detect_card(com_handle, mac_addr)
    if data != None:
        print(bytes(data).decode())

    data = k720.s50_get_card_id(com_handle, mac_addr)
    if data != None and data[0] == ord('P'):
        card_id = data[3:]
        k720.print_packet('Card ID: ', card_id) 

    data = k720.s50_load_sec_key(com_handle, mac_addr, 0x00, k720.KEYA, [0xff, 0xff, 0xff, 0xff, 0xff, 0xff])
    if data != None and data[0] == ord('P'):
        print('Password check successfull')
    else:
        print('Password check failure')
    data = k720.s50_load_sec_key(com_handle, mac_addr, 0x00, k720.KEYA, [0x00, 0xff, 0xff, 0xff, 0xff, 0xff])
    if data != None and data[0] == ord('P'):
        print('Password check successfull')
    else:
        print('Password check failure')

    #############################################################################

    print('== S70 ==')

    data = k720.s70_detect_card(com_handle, mac_addr)
    if data != None:
        print(bytes(data).decode())

    data = k720.s70_get_card_id(com_handle, mac_addr)
    if data != None and data[0] == ord('P'):
        card_id = data[3:]
        k720.print_packet('Card ID: ', card_id) 

    #############################################################################

    print('== UL ==')

    data = k720.ul_detect_card(com_handle, mac_addr)
    if data != None:
        print(bytes(data).decode())

    data = k720.ul_get_card_id(com_handle, mac_addr)
    if data != None and data[0] == ord('P'):
        card_id = data[3:]
        k720.print_packet('Card ID: ', card_id) 

    #############################################################################

    if False:
        packet = k720._send_packet(com_handle, mac_addr, '\x47\x30')
        data = k720._get_packet_data(packet)
        print(bytes(data))

        packet = k720._send_packet(com_handle, mac_addr, '\x48\x30')
        data = k720._get_packet_data(packet)
        print(bytes(data))

def card_positions():
    # Send Card
    # k720.send_cmd(com_handle, mac_addr, 'DC')
    # Recycle
    # k720.send_cmd(com_handle, mac_addr, 'CP')
    
    # Outside Position
    # k720.send_cmd(com_handle, mac_addr, 'FC0')    
    # Take Card Position
    # k720.send_cmd(com_handle, mac_addr, 'FC4')
    # Sensor 2 Position
    # k720.send_cmd(com_handle, mac_addr, 'FC6')
    # Card Read Position
    # k720.send_cmd(com_handle, mac_addr, 'FC7')
    # FrontEnterCard
    k720.send_cmd(com_handle, mac_addr, 'FC8')

    sensor_query = k720.sensor_query(com_handle, mac_addr)
    if sensor_query != None:
        k720.print_packet('', sensor_query) 

def operate():
    # Reset
    k720.send_cmd(com_handle, mac_addr, 'RS')
    while True:
        # Outside Position
        k720.send_cmd(com_handle, mac_addr, 'FC0')
        while True:
            sensor_query = k720.sensor_query(com_handle, mac_addr)
            state = k720.calculate_state(sensor_query)
            if state & k720.CARD_AT_SENSOR_1_POSITION:
                break
            # FrontEnterCard
            k720.send_cmd(com_handle, mac_addr, 'FC8')
            # Card Read Position
            k720.send_cmd(com_handle, mac_addr, 'FC7')

        data = k720.s50_get_card_id(com_handle, mac_addr)
        if data != None and data[0] == ord('P'):
            card_id = data[3:]
            k720.print_packet('Card ID: ', card_id) 

        # Take Card Position
        k720.send_cmd(com_handle, mac_addr, 'FC4')

if __name__ == '__main__':
    com_handle = k720.comm_open(port)
    # test1()
    # card_positions()
    operate()
    k720.comm_close(com_handle)
