package rule

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/cel-go/cel"
)

var (
	rejectedErr = errors.New("rejected by rule")
)

type ActionEnum string
type ChangeTypeEnum string

const (
	ActionEnumScript ActionEnum = "script"
	ActionEnumReject ActionEnum = "reject"
)

const (
	ChangeTypeEnumRequest  ChangeTypeEnum = "request"
	ChangeTypeEnumResponse ChangeTypeEnum = "response"
)

type Rule struct {
	Name           string         `yaml:"name"`
	Change         ChangeTypeEnum `yaml:"change"`
	Enabled        bool           `yaml:"enabled"`
	Rule           string         `yaml:"rule"`
	Action         ActionEnum     `yaml:"action"`
	Import         string         `yaml:"import"`
	Script         string         `yaml:"script"`
	CompiledScript func(*http.Request, *http.Response) error
	CompiledRule   cel.Program
}

func (r *Rule) Check(req *http.Request, res *http.Response) (bool, error) {
	v, _, err := r.CompiledRule.Eval(map[string]any{
		"req":  req,
		"resp": res,
	})
	if err != nil {
		return false, fmt.Errorf("failed to evaluate rule %v", err)
	}

	b, ok := v.Value().(bool)
	if !ok {
		return false, fmt.Errorf("result is not bool %t", b)
	}

	return b, nil
}

func (r *Rule) Apply(req *http.Request, resp *http.Response) error {
	slog.Debug("Applying rule", slog.String("rule", r.Name))

	if r.Action == ActionEnumReject {
		return rejectedErr
	}

	return r.CompiledScript(req, resp)
}
