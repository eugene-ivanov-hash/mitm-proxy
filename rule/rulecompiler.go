package rule

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/ext"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"gopkg.in/yaml.v3"
)

func CompileRules(rulesDir string, envs map[string]string) ([]*Rule, []*Rule, error) {
	i := interp.New(interp.Options{})
	if err := i.Use(stdlib.Symbols); err != nil {
		log.Fatalf("failed to use stdlib: %v", err)
	}

	celEnv, err := NewCelEnv()

	requestRules := make([]*Rule, 0)
	responseRules := make([]*Rule, 0)

	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		return requestRules, responseRules, nil
	}

	index := 1
	err = filepath.Walk(rulesDir, func(path string, f os.FileInfo, err error) error {
		e := func(err error) error {
			return fmt.Errorf("filepath.Walk: %s, %v", path, err)
		}
		if err != nil {
			return e(err)
		}

		if !strings.HasSuffix(f.Name(), ".yaml") && !strings.HasSuffix(f.Name(), ".yml") {
			return nil
		}

		dat, err := os.ReadFile(path)
		if err != nil {
			return e(err)
		}

		ruleFile := &struct {
			Rules   []*Rule `yaml:"rules"`
			Enabled bool    `yaml:"enabled"`
		}{}
		err = yaml.Unmarshal(dat, ruleFile)
		if err != nil {
			return e(err)
		}

		if !ruleFile.Enabled {
			slog.Info("Rule file is disabled", slog.String("file", path))
			return nil
		}

		for _, r := range ruleFile.Rules {
			if !r.Enabled {
				slog.Info("Rule is disabled", slog.String("ruleFile", r.Name))
				continue
			}

			err = compileRule(celEnv, r, envs)
			if err != nil {
				return e(err)
			}

			err = compileScripts(index, i, r, envs)
			if err != nil {
				return e(err)
			}

			index++

			switch r.Change {
			case ChangeTypeEnumRequest:
				requestRules = append(requestRules, r)
			case ChangeTypeEnumResponse:
				responseRules = append(responseRules, r)
			default:
				return fmt.Errorf("unknown change type %s", r.Change)
			}
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return requestRules, responseRules, nil
}

func compileRule(celEnv *cel.Env, rule *Rule, envs map[string]string) error {
	t, err := template.New("rule").Parse(rule.Rule)
	if err != nil {
		return err
	}

	var src bytes.Buffer
	tmplData := map[string]interface{}{
		"Envs": envs,
	}

	err = t.Execute(&src, tmplData)
	if err != nil {
		return err
	}

	slog.Debug("Compiling rule", slog.String("rule name", rule.Name), slog.String("rule", src.String()))

	ast, issues := celEnv.Compile(src.String())
	if issues.Err() != nil {
		return issues.Err()
	}

	checked, iss := celEnv.Check(ast)
	if iss.Err() != nil {
		return iss.Err()
	}

	if !reflect.DeepEqual(checked.OutputType(), cel.BoolType) {
		return fmt.Errorf("expected output type %v, but got %v", cel.BoolType, checked.OutputType())
	}

	prg, err := celEnv.Program(ast)
	if err != nil {
		return err
	}

	rule.CompiledRule = prg

	return nil
}

func compileScripts(index int, i *interp.Interpreter, rule *Rule, envs map[string]string) error {
	packageName := fmt.Sprintf("rule%d", index)
	t := template.New(packageName)

	tmpl := `
	package {{ .PackageName }}

	import (
		"net/http"
		"fmt"
		{{ .Import }}
	)

	func Modify(req *http.Request, resp *http.Response) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("{{ .PackageName }}.Modify err: %v", r)
			}
		}()

		{{ .Script }}
	}`

	t, err := t.Parse(tmpl)
	if err != nil {
		return err
	}

	var src bytes.Buffer
	tmplData := map[string]interface{}{
		"PackageName": packageName,
		"Script":      rule.Script,
		"Import":      rule.Import,
		"Envs":        envs,
	}

	err = t.Execute(&src, tmplData)
	if err != nil {
		return err
	}

	slog.Debug("Compiling script", slog.String("rule name", rule.Name), slog.String("script", src.String()))

	_, err = i.Eval(src.String())
	if err != nil {
		return err
	}

	reqFunc, err := i.Eval(packageName + ".Modify")
	if err != nil {
		return err
	}

	rule.CompiledScript = reqFunc.Interface().(func(req *http.Request, resp *http.Response) error)

	return nil
}

func NewCelEnv() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Variable("req", cel.ObjectType("http.Request")),
		cel.Variable("resp", cel.ObjectType("http.Response")),
		ext.NativeTypes(reflect.TypeOf(http.Request{})),
		ext.NativeTypes(reflect.TypeOf(http.Response{})),
		cel.Function(
			"getBody",
			cel.MemberOverload(
				"req_getBody_string",
				[]*cel.Type{cel.ObjectType("http.Request")},
				cel.StringType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					req, ok := values[0].Value().(*http.Request)
					if !ok {
						return types.NewErr("invalid request type")
					}

					if req.Body == nil || req.Body == http.NoBody {
						return types.String("")
					}

					req.Body = io.NopCloser(ReusableReader(req.Body))

					reqBody, err := io.ReadAll(req.Body)
					if err != nil {
						return types.NewErr("failed to read request body: %v", err)
					}

					return types.String(reqBody)
				}),
			),
		),
		cel.Function(
			"getBody",
			cel.MemberOverload(
				"resp_getBody_string",
				[]*cel.Type{cel.ObjectType("http.Response")},
				cel.StringType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					resp, ok := values[0].Value().(*http.Response)
					if !ok {
						return types.NewErr("invalid request type")
					}

					if resp.Body == nil {
						return types.String("")
					}

					resp.Body = io.NopCloser(ReusableReader(resp.Body))

					respBody, err := io.ReadAll(resp.Body)
					if err != nil {
						return types.NewErr("failed to read request body: %v", err)
					}

					return types.String(respBody)
				}),
			),
		),
	)
}
