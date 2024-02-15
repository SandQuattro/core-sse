package errs

import "fmt"

func Wrap(msg string, err error) error {
	return fmt.Errorf("%s: %w", msg, err)
}

func WrapStack(msg string, err error) error {
	return fmt.Errorf("%s: %+v", msg, err)
}

func WrapIfErr(msg string, err error) error {
	if err == nil {
		return nil
	}
	return Wrap(msg, err)
}

func WrapWithStackIfErr(msg string, err error) error {
	if err == nil {
		return nil
	}
	return WrapStack(msg, err)
}
