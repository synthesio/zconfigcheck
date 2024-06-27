package zconfigcheck

import (
	"fmt"
	"sort"
	"strings"

	"github.com/synthesio/zconfig/v2"
)

// Scope contains information about the defined
// injection and key aliases and their related fields.
type Scope struct {
	Sources map[string][]StructField
	Targets map[string][]StructField
	Keys    map[string][]StructField
	EnvKeys map[string][]string
}

func NewScope() Scope {
	return Scope{
		Sources: make(map[string][]StructField),
		Targets: make(map[string][]StructField),
		Keys:    make(map[string][]StructField),
		EnvKeys: make(map[string][]string),
	}
}

// CheckSource returns any possible issues which would arise if the given
// field is added to the scope as an injection source.
func (i *Scope) CheckSource(field StructField) []string {
	var issues []string
	for _, sourceField := range i.Sources[field.Alias] {
		if sourceField.Pos == field.Pos {
			continue
		}

		issues = append(issues, fmt.Sprintf(
			"inject-as alias '%s' defined by field %s is already used by field %s",
			field.Alias, field.Path, sourceField.Path,
		))
	}

	for _, target := range i.Targets[field.Alias] {
		if field.AssignableTo(target) {
			continue
		}
		issues = append(issues, fmt.Sprintf(
			"injection alias '%s': cannot inject source field '%s' into target field '%s', mismatched types",
			field.Alias, field, target,
		))
	}

	return issues
}

// AddSource adds the given field to the scope as an injection source
func (i *Scope) AddSource(field StructField) {
	i.Sources[field.Alias] = append(i.Sources[field.Alias], field)
}

// CheckTarget returns any possible issues which would arise if the given
// field is added to the scope as an injection target.
func (i *Scope) CheckTarget(field StructField) []string {
	var issues []string

	targets := i.Targets[field.Alias]
	for _, target := range targets {
		if target.Pos == field.Pos || field.CompatibleWith(target) {
			continue
		}

		issues = append(issues, fmt.Sprintf(
			"injection alias '%s': target fields '%s' and '%s' are incompatible, mismatched types",
			field.Alias, field, target,
		))
	}

	for _, src := range i.Sources[field.Alias] {
		if src.AssignableTo(field) {
			continue
		}
		issues = append(issues, fmt.Sprintf(
			"injection alias '%s': target field '%s' cannot be injected with source field '%s', mismatched types",
			field.Alias, field, src,
		))
	}

	return issues
}

// AddTarget adds the given field to the scope as an injection target
func (i *Scope) AddTarget(field StructField) {
	i.Targets[field.Alias] = append(i.Targets[field.Alias], field)
}

// CheckKey returns any possible issues which would arise if the given
// field is added to the scope as a configuration key.
func (i *Scope) CheckKey(field StructField) []string {
	for _, key := range i.Keys[field.Key] {
		if key.Pos == field.Pos {
			continue
		}

		return []string{fmt.Sprintf("key '%s' defined by field '%s' is already used by field '%s'", field.Key, field.Path, key.Path)}
	}

	envKey := zconfig.EnvProvider{}.FormatKey(field.Key)
	if envKey == field.Key {
		return nil
	}

	keys, ok := i.EnvKeys[envKey]
	if !ok {
		return nil
	}

	var issues []string
	for _, key := range keys {
		if key == field.Key {
			continue
		}

		issues = append(issues,
			fmt.Sprintf(
				"key '%s' used by field '%s' and key '%s' used by field %s have the same env format '%s'",
				field.Key,
				field.Path,
				key,
				i.Keys[key][0].Path,
				envKey,
			),
		)
	}

	return issues
}

// AddKey adds the given field to the scope as a configuration target
func (i *Scope) AddKey(field StructField) {
	i.Keys[field.Key] = append(i.Keys[field.Key], field)

	if len(i.Keys[field.Key]) == 1 {
		envKey := zconfig.EnvProvider{}.FormatKey(field.Key)
		i.EnvKeys[envKey] = append(i.EnvKeys[envKey], field.Key)
	}
}

func (i Scope) String() string {
	if len(i.Targets) == 0 && len(i.Sources) == 0 && len(i.Keys) == 0 {
		return "<empty>"
	}

	var parts []string

	if len(i.Targets) > 0 {
		var aliases []string
		for alias := range i.Targets {
			aliases = append(aliases, alias)
		}
		sort.Strings(aliases)
		parts = append(parts, fmt.Sprintf("<inject:%s>", strings.Join(aliases, ",")))
	}

	if len(i.Sources) > 0 {
		var aliases []string
		for alias := range i.Sources {
			aliases = append(aliases, alias)
		}
		sort.Strings(aliases)
		parts = append(parts, fmt.Sprintf("<inject-as:%s>", strings.Join(aliases, ",")))
	}

	if len(i.Keys) > 0 {
		var aliases []string
		for alias := range i.Keys {
			aliases = append(aliases, alias)
		}
		sort.Strings(aliases)
		parts = append(parts, fmt.Sprintf("<keys:%s>", strings.Join(aliases, ",")))
	}

	return strings.Join(parts, " ")
}

// Check returns all Issues caused by the current state of the Scope.
// Unresolved targets, i.e. target aliases for which no source alias is defined, are
// not reported.
func (i Scope) Check() Issues {
	var issues = make(Issues)
	for _, fields := range i.Sources {
		for _, field := range fields {
			issues.Add(field.Pos, i.CheckSource(field)...)
		}
	}
	for _, fields := range i.Targets {
		for _, field := range fields {
			issues.Add(field.Pos, i.CheckTarget(field)...)
		}
	}
	for _, fields := range i.Keys {
		for _, field := range fields {
			issues.Add(field.Pos, i.CheckKey(field)...)
		}
	}

	return issues
}

// UnresolvedTargets returns a map whose keys are all the target aliases
// for which no source aliases are defined.
// The map values are lists of paths where the target alias is used.
func (i Scope) UnresolvedTargets() map[string][]string {
	unresolved := make(map[string][]string)

	for alias, targets := range i.Targets {
		if _, ok := i.Sources[alias]; ok {
			continue
		}

		for _, target := range targets {
			unresolved[alias] = append(unresolved[alias], target.Path)
		}
	}

	return unresolved
}
