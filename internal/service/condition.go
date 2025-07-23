package service

import (
	"errors"
	"strconv"
	"time"

	"github.com/gibgibik/go-lineage2-server/pkg/entity"
	"go.uber.org/zap"
)

func CheckCondition(conditionsCombinator string, conditions []Condition, stat *entity.PlayerStat, party map[uint8]entity.PartyMember, logger *zap.SugaredLogger) (bool, error) {
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
				if !result {
					return false, nil
				}
			}
		case "my_hp":
			if conditionsCombinator == ConditionCombinatorOr {
				result = result || (stat.HP.Percent > 0 && checkOperatorCondition(stat.HP.Percent, cval, condition.Operator))
			} else {
				result = result && stat.HP.Percent > 0 && checkOperatorCondition(stat.HP.Percent, cval, condition.Operator)
				if !result {
					return false, nil
				}
			}
		case "my_mp":
			if conditionsCombinator == ConditionCombinatorOr {
				result = result || (stat.MP.Percent > 0 && checkOperatorCondition(stat.MP.Percent, cval, condition.Operator))
			} else {
				result = result && stat.MP.Percent > 0 && checkOperatorCondition(stat.MP.Percent, cval, condition.Operator)
				if !result {
					return false, nil
				}
			}
		case "since_last_success_target":
			if conditionsCombinator == ConditionCombinatorOr {
				result = result || (checkOperatorCondition(float64(time.Now().UnixMilli()-int64(cval)), float64(stat.Target.HpWasPresentAt), condition.Operator))
			} else {
				result = result && checkOperatorCondition(float64(time.Now().UnixMilli()-int64(cval)), float64(stat.Target.HpWasPresentAt), condition.Operator)
				if !result {
					return false, nil
				}
			}
		case "full_target_hp_unchanged_since":
			if conditionsCombinator == ConditionCombinatorOr {
				result = result || (checkOperatorCondition(float64(time.Now().UnixMilli()-int64(cval)), float64(stat.Target.FullHpUnchangedSince), condition.Operator))
			} else {
				result = result && checkOperatorCondition(float64(time.Now().UnixMilli()-int64(cval)), float64(stat.Target.FullHpUnchangedSince), condition.Operator)
				if !result {
					return false, nil
				}
			}
		case "party_member_hp_1":
			fallthrough
		case "party_member_hp_2":
			fallthrough
		case "party_member_hp_3":
			fallthrough
		case "party_member_hp_4":
			fallthrough
		case "party_member_hp_5":
			fallthrough
		case "party_member_hp_6":
			fallthrough
		case "party_member_hp_7":
			fallthrough
		case "party_member_hp_8":
			s := condition.Field
			memberNum, _ := strconv.ParseUint(s[len(s)-1:], 10, 8)
			memberNum++ //change count from 0 to 1
			if conditionsCombinator == ConditionCombinatorOr {
				if val, ok := party[uint8(memberNum)]; ok {
					result = result || checkOperatorCondition(val.HP.Percent, cval, condition.Operator)
				}
			} else {
				if val, ok := party[uint8(memberNum)]; ok {
					result = result && checkOperatorCondition(val.HP.Percent, cval, condition.Operator)
					if !result {
						return false, nil
					}
				} else {
					result = false
				}
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
