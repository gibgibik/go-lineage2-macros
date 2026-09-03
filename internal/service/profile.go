package service

import (
	"encoding/json"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/gibgibik/go-lineage2-macros/internal/preset"
	"go.uber.org/zap"
)

const (
	ActionAssistPartyMember = "/assistpartymember"
	ActionFollowPartyMember = "/followpartymember"
	ActionAssist            = "/assist"
	ActionAttack            = "/attack"
	ActionTarget            = "/target"
	ActionTargetNext        = "/targetnext"
	ActionDelay             = "/delay"
	ActionPress             = "/press"
	ActionPickup            = "/pickup"
	ActionAITargetNext      = "/aitargetnext"
	ActionStop              = "/stop"
	ActionUnstuck           = "/unstuck"

	ConditionCombinatorAnd = "AND"
	ConditionCombinatorOr  = "OR"
)

type ProfileTemplate struct {
	Items            []ProfilePreset `json:"items"`
	Name             string          `json:"name"`
	PreferredTargets []string        `json:"preferred_targets"`
	AllowedTargets   []string        `json:"allowed_targets"`
}

type ProfilePreset struct {
	Preset   preset.Preset `json:"preset"`
	IsActive bool          `json:"is_active"`
	BatchRun bool          `json:"batch_run"`
}

type ProfileTemplateItem struct {
	Action               string
	Binding              string
	PeriodMilliseconds   int64 `json:"period_milliseconds"`
	DelayMilliseconds    int64 `json:"delay_milliseconds"`
	Additional           string
	Conditions           []preset.Condition
	ConditionsCombinator string `json:"conditions_combinator"`
}

func GetProfileData(profileName string, logger *zap.SugaredLogger) (*ProfileTemplate, error) {
	fileName := getProfilePath(profileName)
	fh, err := os.OpenFile(fileName, os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	buf, err := io.ReadAll(fh)
	if err != nil {
		return nil, err
	}
	var templateBody *ProfileTemplate
	err = json.Unmarshal(buf, &templateBody)
	if err != nil {
		return nil, err
	}
	return templateBody, nil
}

func getProfilePath(profileName string) string {
	reg := regexp.MustCompile("\\W")
	fileName := "var/profiles/" + reg.ReplaceAllString(profileName, "") + ".json" //@todo move to config
	return fileName
}

func SaveProfileData(body io.Reader, logger *zap.SugaredLogger) error {
	inputBody, err := io.ReadAll(body)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	var templateBody ProfileTemplate
	err = json.Unmarshal(inputBody, &templateBody)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	//@todo validation, save as is for now
	fileName := getProfilePath(templateBody.Name)
	tb, err := json.Marshal(templateBody)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	err = os.WriteFile(fileName, tb, 0600)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	logger.Info("profile saved: ", templateBody.Name)
	return nil
}

func GetProfilesList() []string {
	entries, _ := os.ReadDir("var/profiles")
	var result []string
	for _, entry := range entries {
		pieces := strings.Split(entry.Name(), ".")
		result = append(result, pieces[0])
	}

	return result
}
