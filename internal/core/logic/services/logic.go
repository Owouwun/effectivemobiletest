package services

import "time"

func minDate(date1, date2 time.Time) time.Time {
	if date1.Before(date2) {
		return date1
	}
	return date2
}

func maxDate(date1, date2 time.Time) time.Time {
	if date1.After(date2) {
		return date1
	}
	return date2
}

func monthsBetween(startDate, endDate time.Time) int {
	if startDate.After(endDate) {
		startDate, endDate = endDate, startDate
	}

	years := endDate.Year() - startDate.Year()
	months := int(endDate.Month()) - int(startDate.Month())
	return years*12 + months
}
