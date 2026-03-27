package navigation

import "vocoding.net/vocode/v2/apps/daemon/internal/intent"

// Service validates and returns navigation intents for extension execution.
type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) DispatchIntent(nav intent.NavigationIntent) (intent.NavigationIntent, error) {
	if err := intent.ValidateNavigationIntent(nav); err != nil {
		return intent.NavigationIntent{}, err
	}
	return nav, nil
}
