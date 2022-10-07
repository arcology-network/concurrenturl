package common

import "math"

// User access control
const (
	USER_READABLE = iota
	USER_CREATABLE
	USER_UPDATABLE
)

const (
	MaxDepth uint8  = 12
	SYSTEM          = math.MaxInt32
	Root     string = "/"

	CommutativeMeta    uint8 = 100
	CommutativeInt64   uint8 = 101
	CommutativeBalance uint8 = 102

	NoncommutativeInt64  uint8 = 103
	NoncommutativeString uint8 = 104
	NoncommutativeBigint uint8 = 105
	NoncommutativeBytes  uint8 = 106

	VARIATE_TRANSITIONS   uint8 = 0
	INVARIATE_TRANSITIONS uint8 = 1
)
