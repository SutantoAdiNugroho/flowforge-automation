package domain

import "encoding/json"

func ParseDAGDefinition(def JSONB) (*DAGDefinition, error) {
	b, err := json.Marshal(def)
	if err != nil {
		return nil, err
	}
	var dag DAGDefinition
	if err := json.Unmarshal(b, &dag); err != nil {
		return nil, err
	}
	return &dag, nil
}

func TopologicalLevels(steps []DAGStep) ([][]DAGStep, error) {
	stepMap := make(map[string]DAGStep)
	inDegree := make(map[string]int)
	dependents := make(map[string][]string)

	for _, s := range steps {
		stepMap[s.ID] = s
		inDegree[s.ID] = len(s.DependsOn)
		for _, dep := range s.DependsOn {
			dependents[dep] = append(dependents[dep], s.ID)
		}
	}

	var levels [][]DAGStep
	visited := 0

	for visited < len(steps) {
		var level []DAGStep
		for id, deg := range inDegree {
			if deg == 0 {
				level = append(level, stepMap[id])
			}
		}
		if len(level) == 0 {
			break
		}
		for _, s := range level {
			delete(inDegree, s.ID)
			for _, dep := range dependents[s.ID] {
				inDegree[dep]--
			}
		}
		levels = append(levels, level)
		visited += len(level)
	}

	return levels, nil
}
