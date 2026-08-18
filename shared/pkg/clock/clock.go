package clock

import "time"

type RealClock struct{}

func NewRealClock() *RealClock {
	return &RealClock{}
}

func (c *RealClock) Now() time.Time {
	return time.Now().UTC() //UTC это timestamptz? то есть время как у пользователя? как вообще определяется какой у пользователя часовой пояс?
}
