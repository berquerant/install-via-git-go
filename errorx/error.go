package errorx

import "fmt"

func Errorf(err error, format string, a ...any) error {
	return fmt.Errorf("%s %w", fmt.Sprintf(format, a...), err)
}
