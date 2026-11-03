package main

type Solution struct{}

func (s Solution) Solve(inputData any) any {
	m, ok := inputData.(map[string]any)
	if !ok {
		return ""
	}
	name, ok := m["name"].(string)
	if !ok {
		return ""
	}
	return "Hello " + name
}
