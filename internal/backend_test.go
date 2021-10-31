package internal

import (
	"github.com/asecurityteam/rolling"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	linearFan = map[int][]float64{
		0:   {0.0},
		255: {255.0},
	}

	neverStoppingFan = map[int][]float64{
		0:   {50.0},
		50:  {50.0},
		255: {255.0},
	}

	cappedFan = map[int][]float64{
		0:   {0.0},
		1:   {0.0},
		2:   {0.0},
		3:   {0.0},
		4:   {0.0},
		5:   {0.0},
		6:   {20.0},
		200: {200.0},
	}

	cappedNeverStoppingFan = map[int][]float64{
		0:   {50.0},
		50:  {50.0},
		200: {200.0},
	}
)

func createFan(neverStop bool, curveData map[int][]float64) *Fan {
	CurrentConfig.RpmRollingWindowSize = 10

	fan := Fan{
		Config: &FanConfig{
			Id:        "fan1",
			Platform:  "platform",
			Fan:       1,
			NeverStop: neverStop,
			Curve:     "curve",
		},
		FanCurveData: &map[int]*rolling.PointPolicy{},
	}

	AttachFanCurveData(&curveData, &fan)

	return &fan
}

func TestLinearFan(t *testing.T) {
	// GIVEN
	fan := createFan(false, linearFan)

	// WHEN
	startPwm, maxPwm := GetPwmBoundaries(fan)

	// THEN
	assert.Equal(t, 1, startPwm)
	assert.Equal(t, 255, maxPwm)
}

func TestNeverStoppingFan(t *testing.T) {
	// GIVEN
	fan := createFan(false, neverStoppingFan)

	// WHEN
	startPwm, maxPwm := GetPwmBoundaries(fan)

	// THEN
	assert.Equal(t, 0, startPwm)
	assert.Equal(t, 255, maxPwm)
}

func TestCappedFan(t *testing.T) {
	// GIVEN
	fan := createFan(false, cappedFan)

	// WHEN
	startPwm, maxPwm := GetPwmBoundaries(fan)

	// THEN
	assert.Equal(t, 6, startPwm)
	assert.Equal(t, 200, maxPwm)
}

func TestCappedNeverStoppingFan(t *testing.T) {
	// GIVEN
	fan := createFan(false, cappedNeverStoppingFan)

	// WHEN
	startPwm, maxPwm := GetPwmBoundaries(fan)

	// THEN
	assert.Equal(t, 0, startPwm)
	assert.Equal(t, 200, maxPwm)
}

func TestCalculateTargetSpeedLinear(t *testing.T) {
	// GIVEN
	avgTmp := 50000.0
	SensorMap["sensor"] = &Sensor{
		Config: &SensorConfig{
			Id:       "sensor",
			Platform: "platform",
			Index:    0,
		},
		MovingAvg: avgTmp,
	}
	CurveMap["curve"] = &CurveConfig{
		Id:   "curve",
		Type: LinearCurveType,
		Params: LinearCurveConfig{
			Sensor:  "sensor",
			MinTemp: 40,
			MaxTemp: 60,
		},
	}

	fan := createFan(false, linearFan)

	// WHEN
	optimal, err := calculateOptimalPwm(fan)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	// THEN
	assert.Equal(t, 127, optimal)
}

func TestCalculateTargetSpeedNeverStop(t *testing.T) {
	// GIVEN
	avgTmp := 40000.0
	SensorMap["sensor"] = &Sensor{
		Config: &SensorConfig{
			Id:       "sensor",
			Platform: "platform",
			Index:    0,
		},
		MovingAvg: avgTmp,
	}
	CurveMap["curve"] = &CurveConfig{
		Id:   "curve",
		Type: LinearCurveType,
		Params: LinearCurveConfig{
			Sensor:  "sensor",
			MinTemp: 40,
			MaxTemp: 60,
		},
	}

	fan := createFan(true, cappedFan)

	// WHEN
	optimal, err := calculateOptimalPwm(fan)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	target := calculateTargetPwm(fan, 0, optimal)

	// THEN
	assert.Equal(t, 0, optimal)
	assert.Equal(t, fan.StartPwm, target)
}
