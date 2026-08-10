package diagnostics

import (
	"encoding/json"
	"fmt"
)

type Level string

const (
	LevelOff   Level = "off"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

func ParseLevel(s string) (Level, error) {
	switch l := Level(s); l {
	case LevelOff, LevelWarn, LevelError:
		return l, nil
	default:
		return "", fmt.Errorf("invalid level %q (want off|warn|error)", s)
	}
}

type Flag int

const (
	FlagNone Flag = iota
	FlagStrictStringConcat
	FlagStrictCrossTypeCompares
)

func (f Flag) String() string {
	switch f {
	case FlagStrictStringConcat:
		return "strictStringConcat"
	case FlagStrictCrossTypeCompares:
		return "strictCrossTypeCompares"
	}
	return ""
}

func (f Flag) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

type Config struct {
	StrictStringConcat      Level `json:"strictStringConcat"`
	StrictCrossTypeCompares Level `json:"strictCrossTypeCompares"`
}

func DefaultConfig() Config {
	return Config{
		StrictStringConcat:      LevelError,
		StrictCrossTypeCompares: LevelWarn,
	}
}
