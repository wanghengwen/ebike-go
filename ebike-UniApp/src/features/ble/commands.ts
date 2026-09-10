/** BLE command / voice index constants. */
export const BLE_VOICE = {
  START: 3,
  TEMP_LOCK: 7,
  LOCK: 1,
} as const

export const BLE_CMD = {
  PLAY_VOICE: 0x28,
  LOCK: 0x2b,
  UNLOCK: 0x2c,
  GET_DEVICE_INFO: 0x41,
  GET_LAST_BEACON_INFO: 0x42,
  GET_RFID: 0x54,
  GET_GPS: 0x32,
} as const

export type BleCmdCode = (typeof BLE_CMD)[keyof typeof BLE_CMD]
