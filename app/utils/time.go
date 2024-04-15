package utils

import (
	"time"
)

type TimeHelper struct {
	Timezone string
}

func NewTimeHelper(timezone string) (*TimeHelper, error) {
	_, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, err
	}
	return &TimeHelper{
		Timezone: timezone,
	}, nil
}

func (t *TimeHelper) ToISO8601(d time.Time) (string, error) {
	location, err := time.LoadLocation(t.Timezone)
	if err != nil {
		return "", err
	}
	timezoned := d.In(location)
	return timezoned.Format("2006-01-02T15:04:05-0700"), nil
}

func (t *TimeHelper) FromISO8601(d string) (time.Time, error) {
	parsedTime, err := time.Parse("2006-01-02T15:04:05-0700", d)
	if err != nil {
		return time.Time{}, err
	}
	return parsedTime, nil
}

func (t *TimeHelper) NowWithTimezone() (time.Time, error) {
	location, err := time.LoadLocation(t.Timezone)
	if err != nil {
		return time.Time{}, err
	}
	timezoned := time.Now().In(location)
	return timezoned, nil
}

func (t *TimeHelper) NowWithTimezoneISO8601() (string, error) {
	nowTimezoned, err := t.NowWithTimezone()
	if err != nil {
		return "", err
	}
	iso8601, err := t.ToISO8601(nowTimezoned)
	if err != nil {
		return "", err
	}
	return iso8601, nil
}
