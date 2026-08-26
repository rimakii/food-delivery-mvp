package menu

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

//ListRestaurants, GetMenu, UpsertMenuItem, AdjustStock
