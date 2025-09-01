package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Holiday struct {
	Calendar string    `gorm:"primary_key"`
	Date     time.Time `gorm:"primary_key"` // date at 00:00 in Europe/Lisbon
}

func SeedHolidaysPT(db *gorm.DB, years []int) error {
	loc, _ := time.LoadLocation("Europe/Lisbon")
	for _, y := range years {
		// fixed
		fixed := []time.Time{
			time.Date(y, time.January, 1, 0, 0, 0, 0, loc),
			time.Date(y, time.April, 25, 0, 0, 0, 0, loc),
			time.Date(y, time.May, 1, 0, 0, 0, 0, loc),
			time.Date(y, time.June, 10, 0, 0, 0, 0, loc),
			time.Date(y, time.August, 15, 0, 0, 0, 0, loc),
			time.Date(y, time.October, 5, 0, 0, 0, 0, loc),
			time.Date(y, time.November, 1, 0, 0, 0, 0, loc),
			time.Date(y, time.December, 1, 0, 0, 0, 0, loc),
			time.Date(y, time.December, 8, 0, 0, 0, 0, loc),
			time.Date(y, time.December, 25, 0, 0, 0, 0, loc),
		}
		for _, d := range fixed {
			h := Holiday{Calendar: "PT", Date: d}
			_ = db.FirstOrCreate(&h, h).Error
		}
		// movable (Good Friday, Corpus Christi)
		e := easterSunday(y, loc)
		mv := []time.Time{
			e.AddDate(0, 0, -2), // Good Friday
			e.AddDate(0, 0, 60), // Corpus Christi
		}
		for _, d := range mv {
			h := Holiday{Calendar: "PT", Date: d}
			_ = db.FirstOrCreate(&h, h).Error
		}
	}
	return nil
}

func IsHoliday(db *gorm.DB, calendar string, t time.Time) (bool, error) {
	loc, _ := time.LoadLocation("Europe/Lisbon")
	day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	var h Holiday
	err := db.Where("calendar = ? AND date = ?", calendar, day).First(&h).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func easterSunday(y int, loc *time.Location) time.Time {
	a := y % 19
	b := y / 100
	c := y % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := ((h + l - 7*m + 114) % 31) + 1
	return time.Date(y, time.Month(month), day, 0, 0, 0, 0, loc)
}
