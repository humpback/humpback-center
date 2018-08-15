package cluster

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

const (
	eq = iota
	noteq

	// NodeLabelsPrefix is the constraint key prefix for node labels.
	NodeLabelsPrefix = "node.labels."
	// EngineLabelsPrefix is the constraint key prefix for engine labels.
	EngineLabelsPrefix = "engine.labels."
)

var (
	alphaNumeric = regexp.MustCompile(`^(?i)[a-z_][a-z0-9\-_.]+$`)
	// value can be alphanumeric and some special characters. it shouldn't container
	// current or future operators like '>, <, ~', etc.
	valuePattern = regexp.MustCompile(`^(?i)[a-z0-9:\-_\s\.\*\(\)\?\+\[\]\\\^\$\|\/]+$`)
	// operators defines list of accepted operators
	operators = []string{"==", "!="}
)

// Constraint defines a constraint.
type Constraint struct {
	key      string
	operator int
	exp      string
}

// ParseConstraints parses list of constraints.
func ParseConstraints(constraints []string) ([]Constraint, error) {

	exprs := []Constraint{}
	for _, c := range constraints {
		found := false
		// each expr is in the form of "key op value"
		for i, op := range operators {
			if !strings.Contains(c, op) {
				continue
			}
			// split with the op
			parts := strings.SplitN(c, op, 2)

			if len(parts) < 2 {
				return nil, fmt.Errorf("invalid expr: %s", c)
			}

			part0 := strings.TrimSpace(parts[0])
			// validate key
			matched := alphaNumeric.MatchString(part0)
			if matched == false {
				return nil, fmt.Errorf("key '%s' is invalid", part0)
			}

			part1 := strings.TrimSpace(parts[1])

			// validate Value
			matched = valuePattern.MatchString(part1)
			if matched == false {
				return nil, fmt.Errorf("value '%s' is invalid", part1)
			}
			// TODO(dongluochen): revisit requirements to see if globing or regex are useful
			exprs = append(exprs, Constraint{key: part0, operator: i, exp: part1})

			found = true
			break // found an op, move to next entry
		}
		if !found {
			return nil, fmt.Errorf("constraint expected one operator from %s", strings.Join(operators, ", "))
		}
	}
	return exprs, nil
}

// Match checks if the Constraint matches the target strings.
func (c *Constraint) Match(whats ...string) bool {

	var match bool
	// full string match
	for _, what := range whats {
		// case insensitive compare
		if strings.EqualFold(c.exp, what) {
			match = true
			break
		}
	}

	switch c.operator {
	case eq:
		return match
	case noteq:
		return !match
	}
	return false
}

// MatchConstraints returns true if the node satisfies the given constraints.
func MatchConstraints(constraints []Constraint, engine *Engine) bool {

	for _, constraint := range constraints {
		switch {
		case strings.EqualFold(constraint.key, "node.id"):
			if !constraint.Match(engine.ID) {
				return false
			}
		case strings.EqualFold(constraint.key, "node.hostname"):
			// if this node doesn't have hostname
			// it's equivalent to match an empty hostname
			// where '==' would fail, '!=' matches
			if engine.Name == "" {
				if !constraint.Match("") {
					return false
				}
				continue
			}
			if !constraint.Match(engine.Name) {
				return false
			}
		case strings.EqualFold(constraint.key, "node.ip"):
			engineIP := net.ParseIP(engine.IP)
			// single IP address, node.ip == 2001:db8::2
			if ip := net.ParseIP(constraint.exp); ip != nil {
				ipEq := ip.Equal(engineIP)
				if (ipEq && constraint.operator != eq) || (!ipEq && constraint.operator == eq) {
					return false
				}
				continue
			}
			// CIDR subnet, node.ip != 210.8.4.0/24
			if _, subnet, err := net.ParseCIDR(constraint.exp); err == nil {
				within := subnet.Contains(engineIP)
				if (within && constraint.operator != eq) || (!within && constraint.operator == eq) {
					return false
				}
				continue
			}
			// reject constraint with malformed address/network
			return false
		/*
			case strings.EqualFold(constraint.key, "node.role"):
				if !constraint.Match(n.Role.String()) {
					return false
				}
		*/
		case strings.EqualFold(constraint.key, "node.platform.os"):
			if engine.OSType == "" {
				if !constraint.Match("") {
					return false
				}
				continue
			}
			if !constraint.Match(engine.OSType) {
				return false
			}
		case strings.EqualFold(constraint.key, "node.platform.arch"):
			if engine.Architecture == "" {
				if !constraint.Match("") {
					return false
				}
				continue
			}
			if !constraint.Match(engine.Architecture) {
				return false
			}
		// node labels constraint in form like 'node.labels.key==value'
		case len(constraint.key) > len(NodeLabelsPrefix) && strings.EqualFold(constraint.key[:len(NodeLabelsPrefix)], NodeLabelsPrefix):
			if engine.NodeLabels == nil {
				if !constraint.Match("") {
					return false
				}
				continue
			}
			label := constraint.key[len(NodeLabelsPrefix):]
			// label itself is case sensitive
			val := engine.NodeLabels[label]
			if !constraint.Match(val) {
				return false
			}
		// engine labels constraint in form like 'engine.labels.key!=value'
		case len(constraint.key) > len(EngineLabelsPrefix) && strings.EqualFold(constraint.key[:len(EngineLabelsPrefix)], EngineLabelsPrefix):
			if engine.EngineLabels == nil {
				if !constraint.Match("") {
					return false
				}
				continue
			}
			label := constraint.key[len(EngineLabelsPrefix):]
			val := engine.EngineLabels[label]
			if !constraint.Match(val) {
				return false
			}
		default:
			// key doesn't match predefined syntax
			return false
		}
	}
	return true
}
