package preset

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

const (
	ActionAssistPartyMember = "/assistpartymember"
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

type Preset struct {
	Items []Item
	Name  string
}

type Item struct {
	Action               string
	Binding              string
	PeriodMilliseconds   int64 `json:"period_milliseconds"`
	DelayMilliseconds    int64 `json:"delay_milliseconds"`
	Additional           string
	Conditions           []Condition
	ConditionsCombinator string `json:"conditions_combinator"`
}

type Condition struct {
	Id          string `json:"id"`
	Field       string `json:"field"`
	Operator    string `json:"operator"`
	ValueSource string `json:"value_source"`
	Value       string `json:"value"`
}

const (
	PathPrefix = "var/preset/"
)

func GetList(logger *zap.SugaredLogger) ([]Item, error) {
	entries, err := os.ReadDir(PathPrefix)
	if err != nil {
		return nil, err
	}
	result := make([]Item, 0)
	for _, entry := range entries {
		f, err := os.Open(PathPrefix + entry.Name())
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}
		var item Item
		err = json.Unmarshal(b, &item)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, nil
}

func GetProfileData(path string, logger *zap.SugaredLogger) (*Preset, error) {
	pathPieces := strings.SplitN(path, "/", 4)
	if len(pathPieces) < 3 {
		logger.Infof("invalid request", path)
		return nil, errors.New("invalid request")
	}
	fileName := getProfilePath(pathPieces[2])
	fh, err := os.OpenFile(fileName, os.O_RDWR, 0600)
	if errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	defer fh.Close()
	buf, err := io.ReadAll(fh)
	var templateBody *Preset
	err = json.Unmarshal(buf, &templateBody)
	if err != nil {
		return nil, err
	}
	return templateBody, err
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
	var templateBody Preset
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
