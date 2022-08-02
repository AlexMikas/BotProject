package twitchbot

import (
	"time"
)

type User struct {
	currentName string
	oldNames []string
	id int //?????? узнать как на твиче выглядит id

	notes []string
	messages []string

	visitCount int
	viewingTime time.Time
	lastVisit time.Time

	isFollow bool
	followDate time.Time
	unfollowDate time.Time

	isSub bool

	isBot bool
	isStreamer bool
}

func (user *User) AddNewUser() {
	
}

func (user *User) FollowTime() time.Duration {
	if user.isFollow {
		return time.Since(user.followDate)
	} else {
		return 0
	}
}