package edits

import (
	"strings"

	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Apply(params protocol.EditApplyParams) (protocol.EditApplyResult, error) {
	before, after, ok := firstBraceAnchors(params.FileText)
	if !ok {
		return protocol.EditApplyResult{
			Actions: []protocol.EditAction{},
		}, nil
	}

	action := protocol.ReplaceBetweenAnchorsAction{
		Kind: "replace_between_anchors",
		Path: params.ActiveFile,
		Anchor: protocol.Anchor{
			Before: before,
			After:  after,
		},
		NewText: "\n  console.log(\"hi from vocode\");\n",
	}

	return protocol.EditApplyResult{
		Actions: []protocol.EditAction{action},
	}, nil
}

func firstBraceAnchors(fileText string) (before string, after string, ok bool) {
	lines := strings.Split(fileText, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(line, "{") && trimmed != "{" {
			return line, "}", true
		}
	}

	return "", "", false
}
