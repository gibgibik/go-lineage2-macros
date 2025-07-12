package service

import (
	"errors"
	"strconv"
	"time"

	"github.com/gibgibik/go-lineage2-macros/internal/core/entity"
)

func CheckCondition(conditions []Condition, stat *entity.PlayerStat) (bool, error) {
	if stat == nil {
		return false, errors.New("empty player stat, please check server")
	}
	result := true
	for _, condition := range conditions {
		if condition.Value == "" {
			continue
		}
		cval, _ := strconv.ParseFloat(condition.Value, 64)
		switch condition.Field {
		case "target_hp":
			result = result && checkOperatorCondition(stat.Target.HpPercent, cval, condition.Operator)
		case "my_hp":
			result = result && stat.HP.Percent > 0 && checkOperatorCondition(stat.HP.Percent, cval, condition.Operator)
		case "my_mp":
			result = result && stat.MP.Percent > 0 && checkOperatorCondition(stat.MP.Percent, cval, condition.Operator)
		case "since_last_success_target":
			result = result && checkOperatorCondition(float64(stat.Target.HpWasPresentAt), float64(time.Now().Unix()-int64(cval)), condition.Operator)
		}
	}
	return result, nil
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
