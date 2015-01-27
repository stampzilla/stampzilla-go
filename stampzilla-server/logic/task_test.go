package logic

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jonaz/astrotime"

	"code.google.com/p/go-uuid/uuid"
)

func TestTaskSetUuid(t *testing.T) {
	task := &task{}
	task.SetUuid("test")
	if task.Uuid() == "test" {
		return
	}
	t.Errorf("Uuid wrong expected: %s got %s", "test", task.Uuid())
}

func TestTaskRunAndAddActions(t *testing.T) {
	task := &task{}
	actionRunCount := 0
	action := NewRuleActionStub(&actionRunCount)
	task.AddAction(action)
	task.AddAction(action)
	task.Run()
	task.Run()
	if actionRunCount == 4 {
		return
	}
	t.Errorf("actionRunCount wrong expected: %d got %d", 4, actionRunCount)
}
func TestSchedulerCalculateSun(t *testing.T) {

	sunset := astrotime.NextSunset(time.Now(), float64(56.878333), float64(14.809167))
	sunrise := astrotime.NextSunrise(time.Now(), float64(56.878333), float64(14.809167))
	dusk := astrotime.NextDusk(time.Now(), float64(56.878333), float64(14.809167), astrotime.CIVIL_DUSK)
	dawn := astrotime.NextDawn(time.Now(), float64(56.878333), float64(14.809167), astrotime.CIVIL_DAWN)

	when := ""
	task := &task{Name_: "Test", Uuid_: uuid.New()}

	when = task.CalculateSun("* sunset sunset * * *")
	fmt.Println("sunset:", when)
	s := strings.Split(when, " ")
	if s[1] != strconv.Itoa(sunset.Minute()) || s[2] != strconv.Itoa(sunset.Hour()) {
		t.Errorf("Sunset does not match %s %s", s[1], sunset.Minute())
	}

	when = task.CalculateSun("* sunrise sunrise * * *")
	fmt.Println("sunrise:", when)
	s = strings.Split(when, " ")
	if s[1] != strconv.Itoa(sunrise.Minute()) || s[2] != strconv.Itoa(sunrise.Hour()) {
		t.Errorf("Sunrise does not match %s %s", s[1], sunrise.Minute())
	}

	when = task.CalculateSun("* dawn dawn * * *")
	fmt.Println("dawn:", when)
	s = strings.Split(when, " ")
	if s[1] != strconv.Itoa(dawn.Minute()) || s[2] != strconv.Itoa(dawn.Hour()) {
		t.Errorf("Dawn does not match %s %s", s[1], dawn.Minute())
	}

	when = task.CalculateSun("* dusk dusk * * *")
	fmt.Println("dusk:", when)
	s = strings.Split(when, " ")
	if s[1] != strconv.Itoa(dusk.Minute()) || s[2] != strconv.Itoa(dusk.Hour()) {
		t.Errorf("Dusk does not match %s %s", s[1], dusk.Minute())
	}

	//t.Errorf("actionRunCount wrong expected cron to have ran 4 times got %d", actionRunCount)
}
