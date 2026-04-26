package appsetup

import userModule "pleco-api/internal/modules/user"

type accessTokenVersionAdapter struct {
	repo userModule.Repository
}

func (a accessTokenVersionAdapter) AccessTokenVersionForUser(userID uint) (uint, error) {
	u, err := a.repo.FindByID(userID)
	if err != nil {
		return 0, err
	}
	return u.AccessTokenVersion, nil
}
