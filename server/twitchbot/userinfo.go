package twitchbot

import "time"

type User struct {
	currentName string
	oldNames []string
	id int //?????? узнать как на твиче выглядит id

	notes []string
	messages []string

	visitCount int
	viewingTime time.Duration
	lastVisit time.Duration

	isFollow bool
	followDate time.Duration // TODO:: Func FollowTime
	unfollowDate time.Duration

	isSub bool

	isBot bool
	isStreamer bool
}
