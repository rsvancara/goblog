package filters

import (
	"errors"
	"math/rand"
	"time"

	"github.com/flosch/go-humanize"
	"github.com/flosch/pongo2"
)

func init() {

	rand.Seed(time.Now().UTC().UnixNano())

	pongo2.RegisterFilter("timeuntil", filterTimeuntilTimesince)
	pongo2.RegisterFilter("timesince", filterTimeuntilTimesince)
	pongo2.RegisterFilter("naturaltime", filterTimeuntilTimesince)
	pongo2.RegisterFilter("naturalday", filterNaturalday)
}

func filterTimeuntilTimesince(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	basetime, isTime := in.Interface().(time.Time)
	if !isTime {
		return nil, &pongo2.Error{
			Sender:    "filter:timeuntil/timesince",
			OrigError: errors.New("time-value is not a time.Time-instance"),
		}
	}
	var paramtime time.Time
	if !param.IsNil() {
		paramtime, isTime = param.Interface().(time.Time)
		if !isTime {
			return nil, &pongo2.Error{
				Sender:    "filter:timeuntil/timesince",
				OrigError: errors.New("time-parameter is not a time.Time-instance"),
			}
		}
	} else {
		paramtime = time.Now()
	}

	return pongo2.AsValue(humanize.TimeDuration(basetime.Sub(paramtime))), nil
}

func filterNaturalday(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	basetime, isTime := in.Interface().(time.Time)
	if !isTime {
		return nil, &pongo2.Error{
			Sender:    "filter:naturalday",
			OrigError: errors.New("naturalday-value is not a time.Time-instance"),
		}
	}

	var referenceTime time.Time
	if !param.IsNil() {
		referenceTime, isTime = param.Interface().(time.Time)
		if !isTime {
			return nil, &pongo2.Error{
				Sender:    "filter:naturalday",
				OrigError: errors.New("naturalday-parameter is not a time.Time-instance"),
			}
		}
	} else {
		referenceTime = time.Now()
	}

	d := referenceTime.Sub(basetime) / time.Hour

	switch {
	case d >= 0 && d < 24:
		// Today
		return pongo2.AsValue("today"), nil
	case d >= 24:
		return pongo2.AsValue("yesterday"), nil
	case d < 0 && d >= -24:
		return pongo2.AsValue("tomorrow"), nil
	}

	// Default behaviour
	return pongo2.ApplyFilter("naturaltime", in, param)
}
