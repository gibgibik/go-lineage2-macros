package entity

type PlayerStat struct {
	CP struct {
		Percent    float64
		LastUpdate int64
	}
	HP struct {
		Percent    float64
		LastUpdate int64
	}
	MP struct {
		Percent    float64
		LastUpdate int64
	}
	EXP struct {
		Percent    float64
		LastUpdate int64
	}
	Target struct {
		HpPercent            float64
		LastUpdate           int64
		HpWasPresentAt       int64
		FullHpUnchangedSince int64
	}
}

const (
	Hp = "HP"
	Mp = "MP"
)
