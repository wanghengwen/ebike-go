package service

// Part check types — mirrors Java PartCheckTypEnum.
const (
	PartCheckUseCar         = 1
	PartCheckReturnCar      = 2
	PartCheckRidingCar      = 3
	PartCheckBlueUseCar     = 4
	PartCheckBlueReturnCar  = 5
	PartCheckGetPart        = 6
	PartCheckFocusReturnCar = 7
	PartCheckAutoReturnCar  = 8
	PartCheckPaybackReturn  = 9
)

const (
	codeInvokeFeignFail = "13011"
	codeDeviceTimeout   = "17012"
)

const (
	effectHasPart = 0
)

const (
	precisePartsCSV = "rfid,kickstand,beacon"
	assertPartsCSV  = "helmet"
)

func isReturnCarPartCheck(checkType int) bool {
	switch checkType {
	case PartCheckReturnCar, PartCheckFocusReturnCar, PartCheckAutoReturnCar, PartCheckBlueReturnCar, PartCheckPaybackReturn:
		return true
	default:
		return false
	}
}

func isBluePartCheck(checkType int) bool {
	return checkType == PartCheckBlueReturnCar || checkType == PartCheckBlueUseCar
}
