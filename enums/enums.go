package enums

// VERSIONS KLOVERCLOUD available versions
type VERSIONS string

const (
	V0_0_1_BETA = VERSIONS("v0.0.1-beta")
	LATEST      = VERSIONS("v0.0.1-beta")
)

// DATABASE_OPTION supported databases
type DATABASE_OPTION string

const (
	MONGO   = DATABASE_OPTION("MONGO")
	DEFAULT = DATABASE_OPTION("MONGO")
)
