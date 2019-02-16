package main

import "fmt"

func boolDiff(target interface{}, current interface{}) (diff bool, value bool, err error) {
	value, ok1 := target.(bool)
	currentBool, ok2 := current.(bool)

	if !ok1 {
		return false, false, fmt.Errorf("Could not cast target value to bool")
	}

	diff = !ok2 || value != currentBool
	return
}

func scalingDiff(target interface{}, current interface{}) (diff bool, value float64, err error) {
	value, ok1 := target.(float64)
	currentFloat, ok2 := current.(float64)

	if !ok1 {
		return false, 0, fmt.Errorf("Could not cast target value to bool")
	}

	diff = !ok2 || value != currentFloat
	return
}
