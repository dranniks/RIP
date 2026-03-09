package identity

import "sync"

type Actor struct {
	ID    uint
	Login string
	Role  string
}

type Users struct {
	Creator   Actor
	Moderator Actor
}

const (
	creatorID      uint = 1
	creatorLogin        = "xrf_creator"
	moderatorID    uint = 2
	moderatorLogin      = "xrf_moderator"
)

var (
	usersOnce sync.Once
	usersRef  *Users
)

// CurrentUsers returns singleton of fixed users used before auth lab.
func CurrentUsers() *Users {
	usersOnce.Do(func() {
		usersRef = &Users{
			Creator: Actor{
				ID:    creatorID,
				Login: creatorLogin,
				Role:  "creator",
			},
			Moderator: Actor{
				ID:    moderatorID,
				Login: moderatorLogin,
				Role:  "moderator",
			},
		}
	})

	return usersRef
}

func CreatorID() uint {
	return CurrentUsers().Creator.ID
}

func ModeratorID() uint {
	return CurrentUsers().Moderator.ID
}
