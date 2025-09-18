package services

import "time"

func minDate(date1 time.Time, date2 time.Time) time.Time {
	if date1.Before(date2) {
		return date1
	}
	return date2
}
func maxDate(date1 time.Time, date2 time.Time) time.Time {
	if date1.After(date2) {
		return date1
	}
	return date2
}
func monthDiff(date1 time.Time, date2 time.Time) int {
	years := date2.Year() - date1.Year()
	months := int(date2.Month()) - int(date1.Month())
	return years*12 + months
}
