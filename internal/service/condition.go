package service

import (
	"strconv"
	"time"

	"github.com/gibgibik/go-lineage2-macros/internal/core/entity"
)

func CheckCondition(conditions []Condition, stat *entity.PlayerStat) bool {
	for _, condition := range conditions {
		if condition.Value == "" {
			continue
		}
		cval, _ := strconv.ParseFloat(condition.Value, 64)
		switch condition.Field {
		case "target_hp":
			return stat.Target.HpWasPresentAt > 0 && checkOperatorCondition(stat.Target.HpPercent, cval, condition.Operator)
		case "my_hp":
			return stat.HP.Percent > 0 && checkOperatorCondition(stat.HP.Percent, cval, condition.Operator)
		case "my_mp":
			return stat.MP.Percent > 0 && checkOperatorCondition(stat.MP.Percent, cval, condition.Operator)
		case "since_last_success_target":
			return checkOperatorCondition(float64(stat.Target.HpWasPresentAt), float64(time.Now().Unix()-int64(cval)), condition.Operator)
		}
	}
	return true
}

func checkOperatorCondition(item float64, item2 float64, operator string) bool {
	switch operator {
	case ">":
		return item > item2
	case "=":
		return item == item2
	case "<":
		return item < item2
	default:
		panic("unsupported operator: " + operator)
	}
}
