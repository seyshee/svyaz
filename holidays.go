package main

import "time"

type occasion struct {
	name, day    string
	hue          int
	professional bool
}

func holiday(day time.Time) obj {
	y := day.Year()
	zone := day.Location()
	date := func(month time.Month, n int) time.Time { return time.Date(y, month, n, 12, 0, 0, 0, zone) }
	weekday := func(m time.Month, w time.Weekday, n int) string {
		start := date(m, 1)
		if n < 0 {
			end := start.AddDate(0, 1, -1)
			return end.AddDate(0, 0, -(int(end.Weekday()-w)+7)%7).Format("01-02")
		}
		return start.AddDate(0, 0, (int(w-start.Weekday())+7)%7+7*(n-1)).Format("01-02")
	}
	d := (19*(y%19) + 15) % 30
	e := (2*(y%4) + 4*(y%7) - d + 34) % 7
	v := d + e + 114
	easter := date(time.Month(v/31), v%31+1).AddDate(0, 0, 13)
	list := []occasion{
		{"Старый Новый год", "01-14", 200, false},
		{"Татьянин день", "01-25", 215, false},
		{"День святого Валентина", "02-14", 345, false},
		{"23 Февраля", "02-23", 140, false},
		{"8 Марта", "03-08", 330, false},
		{"День космонавтики", "04-12", 235, false},
		{"Субботник", weekday(4, time.Saturday, -1), 110, false},
		{"Пасха", easter.Format("01-02"), 45, false},
		{"Весна и Труд", "05-01", 120, false},
		{"День Победы", "05-09", 15, false},
		{"Последний звонок", "05-25", 50, false},
		{"День защиты детей", "06-01", 190, false},
		{"День России", "06-12", 220, false},
		{"День медика", weekday(6, time.Sunday, 3), 175, true},
		{"Иван Купала", "07-07", 95, false},
		{"День семьи, любви и верности", "07-08", 10, false},
		{"День строителя", weekday(8, time.Sunday, 2), 35, true},
		{"День знаний", "09-01", 260, false},
		{"День нефтяника", weekday(9, time.Sunday, 1), 30, true},
		{"День города Москвы", weekday(9, time.Saturday, 2), 0, false},
		{"День программиста", date(1, 256).Format("01-02"), 150, true},
		{"День учителя", "10-05", 40, false},
		{"Хэллоуин", "10-31", 25, false},
		{"День народного единства", "11-04", 5, false},
		{"День бухгалтера", "11-21", 160, true},
		{"День матери", weekday(11, time.Sunday, -1), 340, false},
		{"День банковского работника", "12-02", 140, true},
		{"День юриста", "12-03", 230, true},
		{"День энергетика", "12-22", 50, true},
		{"Новый год", "12-31", 0, false},
	}
	cny := map[int]string{2027: "02-06", 2028: "01-26", 2029: "02-13", 2030: "02-03", 2031: "01-23"}
	list = append(list, occasion{"Китайский Новый год", cny[y], 0, false})
	for _, h := range list {
		if h.day == day.Format("01-02") {
			v := obj{"name": h.name, "from": day.Format(time.DateOnly), "to": day.Format(time.DateOnly), "hue": h.hue}
			if h.professional {
				v["kind"] = "проф"
			}
			return v
		}
	}
	from := easter.AddDate(0, 0, -55).Format(time.DateOnly)
	to := easter.AddDate(0, 0, -49).Format(time.DateOnly)
	if x := day.Format(time.DateOnly); x >= from && x <= to {
		return obj{"name": "Масленица", "from": from, "to": to, "hue": 40}
	}
	return nil
}
func today() time.Time { return time.Now().In(time.FixedZone("MSK", 10800)) }
