package utils

func UntilFirstError(fns []ErrFn) error {
	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}
