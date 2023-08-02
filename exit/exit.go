package exit

import "os"

const (
	CodeSuccess = 0
	CodeFailure = 1
)

func Exit(code int) {
	os.Exit(code)
}

func Succeed() {
	Exit(CodeSuccess)
}

func Fail() {
	Exit(CodeFailure)
}
