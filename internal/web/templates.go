package web

import (
	"text/template"
	"text/template/parse"

	edlib "github.com/hbollon/go-edlib"
)

func collectFields(tmpl *template.Template) []string {
	fields := make([]string, 0)
	seen := make(map[string]struct{})
	collectNode(tmpl.Root, seen, &fields)
	return fields
}

func collectNode(node parse.Node, seen map[string]struct{}, fields *[]string) {
	if node == nil {
		return
	}

	switch node := node.(type) {
	case *parse.ListNode:
		if node == nil {
			return
		}
		for _, child := range node.Nodes {
			collectNode(child, seen, fields)
		}
	case *parse.ActionNode:
		collectPipe(node.Pipe, seen, fields)
	case *parse.IfNode:
		collectBranch(&node.BranchNode, seen, fields)
	case *parse.RangeNode:
		collectBranch(&node.BranchNode, seen, fields)
	case *parse.WithNode:
		collectBranch(&node.BranchNode, seen, fields)
	case *parse.BranchNode:
		collectBranch(node, seen, fields)
	case *parse.PipeNode:
		collectPipe(node, seen, fields)
	case *parse.CommandNode:
		for _, arg := range node.Args {
			collectNode(arg, seen, fields)
		}
	case *parse.ChainNode:
		collectNode(node.Node, seen, fields)
	case *parse.FieldNode:
		if len(node.Ident) > 0 {
			addField(node.Ident[0], seen, fields)
		}
	}
}

func collectBranch(branch *parse.BranchNode, seen map[string]struct{}, fields *[]string) {
	collectPipe(branch.Pipe, seen, fields)
	collectNode(branch.List, seen, fields)
	collectNode(branch.ElseList, seen, fields)
}

func collectPipe(pipe *parse.PipeNode, seen map[string]struct{}, fields *[]string) {
	if pipe == nil {
		return
	}
	for _, command := range pipe.Cmds {
		collectNode(command, seen, fields)
	}
}

func addField(field string, seen map[string]struct{}, fields *[]string) {
	if _, ok := seen[field]; ok {
		return
	}
	seen[field] = struct{}{}
	*fields = append(*fields, field)
}

func SuggestFields(field string, candidates []string) []string {
	if len(candidates) == 0 {
		return nil
	}

	suggestions, err := edlib.FuzzySearchSet(field, candidates, 3, edlib.Levenshtein)
	if err != nil {
		return nil
	}
	return suggestions
}
