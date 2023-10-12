package errorx

import "fmt"

func Errorf(err error, format string, a ...any) error {
	return fmt.Errorf("%s %w", fmt.Sprintf(format, a...), err)
}

func Serial[S ~[]T, T any](list S, errFunc func(T) error) error {
	for _, x := range list {
		if err := errFunc(x); err != nil {
			return err
		}
	}
	return nil
}
