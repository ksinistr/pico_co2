package ens160

// DefaultAddress is the default I2C address for the ENS160.
const DefaultAddress = 0x53

// Registers.
const (
	regPartID   = 0x00
	regOpMode   = 0x10
	regConfig   = 0x11
	regCommand  = 0x12
	regTempIn   = 0x13
	regHumIn    = 0x15
	regStatus   = 0x20
	regAQI      = 0x21
	regTVOC     = 0x22
	regECO2     = 0x24
	regDataT    = 0x30
	regDataRH   = 0x32
	regMISR     = 0x38
	regGPRWrite = 0x40
	regGPRRead  = 0x48
)

// Operating modes.
const (
	ModeDeepSleep = 0x00
	ModeIdle      = 0x01
	ModeStandard  = 0x02
	ModeReset     = 0xF0
)

// Status register bits.
const (
	statusSTATAS = 1 << 7
	statusSTATER = 1 << 6

	statusValidityMask  = 0x0C
	statusValidityShift = 2

	statusNEWDAT = 1 << 1
	statusNEWGPR = 1 << 0
)

// Validity flags.
const (
	ValidityNormalOperation     = 0x00
	ValidityWarmUpPhase         = 0x01 // need ~3 minutes until valid data
	ValidityInitialStartUpPhase = 0x02 // need ~1 hour until valid data
	ValidityInvalidOutput       = 0x03
)

// Commands.
const (
	cmdNOP       = 0x00
	cmdGetAppVer = 0x0E
	cmdClrGPR    = 0xCC
)
