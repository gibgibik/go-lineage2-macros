package entity

type PlayerStat struct {
	CP struct {
		Value      string
		LastUpdate int64
	}
	HP struct {
		Value      string
		LastUpdate int64
	}
	MP struct {
		Value      string
		LastUpdate int64
	}
	EXP struct {
		Value      string
		LastUpdate int64
	}
	Target struct {
		HpPercent  float64
		LastUpdate int64
	}
}
