package util

import (
	"time"

	"github.com/koding/kite"
	"github.com/koding/kite/protocol"
)

func IrcDial(local *kite.Kite) (*kite.Client, error) {
	kites, err := local.GetKites(&protocol.KontrolQuery{
		Username:    local.Config.Username,
		Environment: local.Config.Environment,
		Name:        "irc",
	})

	if err != nil {
		return nil, err
	}

	var irc *kite.Client

	// take first working kite
	for _, kite := range kites {
		if err = kite.DialTimeout(10 * time.Second); err != nil {
			continue
		}

		irc = kite
		break
	}

	if err != nil {
		return nil, err
	}

	return irc, nil
}

func IrcRegister(irc *kite.Client) error {
	_, err := irc.Tell("irc.register", nil)
	return err
}
