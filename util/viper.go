package util

import (
	"fmt"

	"github.com/spf13/viper"
)

type OptionError string

func (o OptionError) Error() string {
	return fmt.Sprintf("option %q is not set", string(o))
}

func RequiredOptions(conf *viper.Viper, options ...string) error {
	for _, o := range options {
		if !conf.IsSet(o) {
			return OptionError(o)
		}
	}

	return nil
}
