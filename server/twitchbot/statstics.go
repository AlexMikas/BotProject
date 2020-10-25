package twitchbot

import (
	"time"
)

// Структура, которая будет постоянно обновляться.
// Статистику можно записывать... Куда?
type StreamStat struct {
	streamName string
	category string

	startTime time.Duration
	endTime time.Duration

	newFollow int
	newSub int

	unfollowCount int
	unSub int

	viewersCount int
	viewerList []User
}