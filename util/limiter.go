package util

import "golang.org/x/time/rate"

var RateLimiter = rate.NewLimiter(6000, 150)
