package container

import (
	"errors"
	"fmt"
	"sort"
)

type ContainerGroupDepsNode struct {
	group string
	level int
	deps  []*ContainerGroupDepsNode
}

type ContainerGroupDeps struct {
	all map[string][]string
}

func NewContainerGroupDeps() *ContainerGroupDeps {
	return &ContainerGroupDeps{
		all: map[string][]string{},
	}
}

func (cpd *ContainerGroupDeps) AddDeps(group string, deps []string) {
	origDeps := func(cpd *ContainerGroupDeps) []string {
		if ori, ok := cpd.all[group]; ok {
			return ori
		} else {
			return []string{}
		}
	}(cpd)
	origDeps = append(origDeps, deps...)
	cpd.all[group] = origDeps
}

func (cpd *ContainerGroupDeps) CreateNode(group string, deps []string, nodeContainer map[string]*ContainerGroupDepsNode) (*ContainerGroupDepsNode, error) {
	node, ok := nodeContainer[group]
	if ok {
		return node, nil
	}
	node = &ContainerGroupDepsNode{
		group: group,
		deps:  []*ContainerGroupDepsNode{},
		level: 0,
	}
	nodeContainer[group] = node
	for _, dGroup := range deps {
		dgDeps, ok := cpd.all[dGroup]
		if !ok {
			return node, errors.New(fmt.Sprintf("Dependancy group(%s) with '%s' is not exists.", dGroup, group))
		}
		dNode, dnErr := cpd.CreateNode(dGroup, dgDeps, nodeContainer)
		if dnErr != nil {
			return node, dnErr
		}
		node.deps = append(node.deps, dNode)
	}
	return node, nil
}

func (cpd *ContainerGroupDeps) SetLevel(node *ContainerGroupDepsNode, lvl int, cycle map[string]bool) {

	if lvl <= node.level {
		return
	}
	node.level = lvl
	if _, ok := cycle[node.group]; ok {
		panic("Cycle detected, when set level for group: " + node.group)
	}
	cycle[node.group] = true
	for _, dn := range node.deps {
		cpd.SetLevel(dn, lvl+1, cycle)
	}
}

// visit from bottom to top
func (cpd *ContainerGroupDeps) VisitByLevel(visit func(lvl int, groups []string) error, continueOnError bool) error {
	nodeContainer := map[string]*ContainerGroupDepsNode{}
	for group, deps := range cpd.all {
		_, err := cpd.CreateNode(group, deps, nodeContainer)
		if err != nil {
			return errors.New(fmt.Sprintf("Node create(group:%s, deps:%+v) error:%s", group, deps, err.Error()))
		}
	}

	for _, node := range nodeContainer {
		cycleDetect := map[string]bool{}
		cpd.SetLevel(node, 1, cycleDetect)
	}

	nodesByLevel := map[int][]string{}
	for _, node := range nodeContainer {
		lvl := node.level
		if nodes, ok := nodesByLevel[lvl]; ok {
			nodes = append(nodes, node.group)
			nodesByLevel[lvl] = nodes
		} else {
			nodesByLevel[lvl] = []string{node.group}
		}
	}

	sortedLevels := []int{}
	for lvl, _ := range nodesByLevel {
		sortedLevels = append(sortedLevels, lvl)
	}
	sort.Ints(sortedLevels)
	for idx := len(sortedLevels) - 1; idx >= 0; idx-- {
		lvl := sortedLevels[idx]
		deps, _ := nodesByLevel[lvl]
		if err := visit(lvl, deps); err != nil {
			fmt.Printf("Visit by level failed. Level:%d, Deps:%+v, Continue:%t, Error:%s\n", lvl, deps, continueOnError, err.Error())
			if !continueOnError {
				return err
			}
		}
	}
	return nil
}

func (cpd *ContainerGroupDeps) List() []string {
	groups := []string{}
	cpd.VisitByLevel(
		func(lvl int, gs []string) error {
			groups = append(groups, gs...)
			return nil
		},
		true)
	return groups
}
