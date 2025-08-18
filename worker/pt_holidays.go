package worker

import "time"

// Feriados fixos nacionais
func fixedPT(y int) map[time.Time]struct{} {
    loc, _ := time.LoadLocation("Europe/Lisbon")
    m := map[time.Time]struct{}{}
    add := func(mm time.Month, dd int) { m[time.Date(y, mm, dd, 0,0,0,0, loc)] = struct{}{} }
    add(time.January,1); add(time.April,25); add(time.May,1); add(time.June,10)
    add(time.August,15); add(time.October,5); add(time.November,1); add(time.December,1)
    add(time.December,8); add(time.December,25)
    return m
}

func easterSunday(y int) time.Time { // algoritmo de Butcher
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
    loc, _ := time.LoadLocation("Europe/Lisbon")
    return time.Date(y, time.Month(month), day, 0, 0, 0, 0, loc)
}

func movablePT(y int) map[time.Time]struct{} {
    m := map[time.Time]struct{}{}
    easter := easterSunday(y)
    goodFriday := easter.AddDate(0,0,-2)
    corpusChristi := easter.AddDate(0,0,60)
    m[goodFriday] = struct{}{}
    m[corpusChristi] = struct{}{}
    return m
}

func isHolidayPT(t time.Time) bool {
    y := t.Year()
    all := fixedPT(y)
    for d := range movablePT(y) { all[d] = struct{}{} }
    _, ok := all[time.Date(y, t.Month(), t.Day(), 0,0,0,0, t.Location())]
    return ok
}
