package service

import (
	"errors"
	"strconv"
	"time"

	"github.com/gibgibik/go-lineage2-macros/internal/core/entity"
)

func CheckCondition(conditionsCombinator string, conditions []Condition, stat *entity.PlayerStat) (bool, error) {
	if stat == nil {
		return false, errors.New("empty player stat, please check server")
	}
	result := conditionsCombinator != ConditionCombinatorOr
	for _, condition := range conditions {
		if condition.Value == "" {
			continue
		}
		cval, _ := strconv.ParseFloat(condition.Value, 64)
		switch condition.Field {
		case "target_hp":
			if conditionsCombinator == ConditionCombinatorOr {
				result = result || checkOperatorCondition(stat.Target.HpPercent, cval, condition.Operator)
			} else {
				result = result && checkOperatorCondition(stat.Target.HpPercent, cval, condition.Operator)
			}
		case "my_hp":
			if conditionsCombinator == ConditionCombinatorOr {
				result = result || (stat.HP.Percent > 0 && checkOperatorCondition(stat.HP.Percent, cval, condition.Operator))

			} else {
				result = result && stat.HP.Percent > 0 && checkOperatorCondition(stat.HP.Percent, cval, condition.Operator)
			}
		case "my_mp":
			if conditionsCombinator == ConditionCombinatorOr {
				result = result || (stat.MP.Percent > 0 && checkOperatorCondition(stat.MP.Percent, cval, condition.Operator))
			} else {
				result = result && stat.MP.Percent > 0 && checkOperatorCondition(stat.MP.Percent, cval, condition.Operator)
			}
		case "since_last_success_target":
			if conditionsCombinator == ConditionCombinatorOr {
				result = result || (checkOperatorCondition(float64(time.Now().Unix()-int64(cval)), float64(stat.Target.HpWasPresentAt), condition.Operator))
			} else {
				result = result && checkOperatorCondition(float64(time.Now().Unix()-int64(cval)), float64(stat.Target.HpWasPresentAt), condition.Operator)
			}
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
