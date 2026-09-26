package config

const (
	PORT            = ":3000"
	PROTOCOL        = "tcp"
	MAX_CONNECTIONS = 20000
)

type PersistenceConfiguration struct {
	Directory  string
	DBFileName string
}

var Persistence = PersistenceConfiguration{}
