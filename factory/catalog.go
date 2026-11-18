// Package factory is the error catalog: define every business error once (kind + code + message template)
// and produce it anywhere.
//
//	var RoleErrors = struct{ SystemRole factory.ErrorSpec }{
//	    SystemRole: factory.Conflict("system_role", "system role '{code}' cannot be deleted"),
//	}
//	RoleErrors.SystemRole.New(factory.P{"code": "ADMIN"})   // DomainError (a value)
//	factory.Fail[Role](RoleErrors.SystemRole, factory.P{...}) // failed Result
//
// Codes and messages may contain {placeholders}. Every spec registers itself so the i18n layer (and its
// completeness test) can enumerate the templates.
package factory

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/IBM/fp-go/v2/either"

	"github.com/solomanpan/kimoj-core/exception"
)

// P are the parameters of an error.
type P map[string]any

type ErrorSpec struct {
	Kind     exception.Kind
	Code     string
	Template string
}

var (
	mu       sync.RWMutex
	registry = map[string]ErrorSpec{}
)

func register(kind exception.Kind, code, template string) ErrorSpec {
	spec := ErrorSpec{Kind: kind, Code: code, Template: template}
	mu.Lock()
	defer mu.Unlock()
	registry[kind.String()+"|"+code+"|"+template] = spec
	return spec
}

func Validation(code, template string) ErrorSpec {
	return register(exception.Validation, code, template)
}
func NotFound(code, template string) ErrorSpec  { return register(exception.NotFound, code, template) }
func Conflict(code, template string) ErrorSpec  { return register(exception.Conflict, code, template) }
func Forbidden(code, template string) ErrorSpec { return register(exception.Forbidden, code, template) }

// All lists every registered spec, ordered by template.
func All() []ErrorSpec {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]ErrorSpec, 0, len(registry))
	for _, s := range registry {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Template < out[j].Template })
	return out
}

var placeholder = regexp.MustCompile(`\{(\w+)\}`)

// Placeholders are the {names} used by the code and the template.
func (s ErrorSpec) Placeholders() []string {
	seen := map[string]bool{}
	var names []string
	for _, m := range placeholder.FindAllStringSubmatch(s.Code+" "+s.Template, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			names = append(names, m[1])
		}
	}
	sort.Strings(names)
	return names
}

// Format fills {placeholders} of a template.
func Format(template string, params P) string {
	return placeholder.ReplaceAllStringFunc(template, func(m string) string {
		if v, ok := params[m[1:len(m)-1]]; ok {
			return fmt.Sprint(v)
		}
		return m
	})
}

// New builds the DomainError; missing parameters are a programming error and panic.
func (s ErrorSpec) New(params P) exception.DomainError {
	for _, n := range s.Placeholders() {
		if _, ok := params[n]; !ok {
			panic(fmt.Sprintf("error '%s' needs %s", s.Code, strings.Join(s.Placeholders(), ", ")))
		}
	}
	return exception.DomainError{Kind: s.Kind, Code: Format(s.Code, params), Message: Format(s.Template, params)}
}

// Fail produces the spec as a failed Result.
func Fail[A any](s ErrorSpec, params P) either.Either[exception.DomainError, A] {
	return either.Left[A](s.New(params))
}
