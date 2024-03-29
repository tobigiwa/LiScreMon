package store

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type ScreenType string

const (
	Active  ScreenType = "active"
	Passive ScreenType = "passive"
)

// ScreenTime represents the time spent on a particular app.
type ScreenTime struct {
	AppName string
	Type    ScreenType
	Time    float64
}

type dailyAppScreenTime map[string]float64

type appInfo struct {
	AppName           string
	Icon              []byte
	ActiveScreenStats dailyAppScreenTime
	PassiveScreenStats dailyAppScreenTime
}

func (ap *appInfo) serialize() ([]byte, error) {
	var res bytes.Buffer
	encoded := gob.NewEncoder(&res)
	if err := encoded.Encode(ap); err != nil {
		return nil, fmt.Errorf("%v:%w", err, ErrSerilization)
	}
	return res.Bytes(), nil
}

func (ap *appInfo) deserialize(data []byte) error {
	decoded := gob.NewDecoder(bytes.NewReader(data))
	if err := decoded.Decode(ap); err != nil {
		return fmt.Errorf("%v:%w", err, ErrDeserilization)
	}
	return nil
}
