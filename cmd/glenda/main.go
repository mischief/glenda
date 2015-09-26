package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/mischief/glenda/util"

	"github.com/koding/kite"
	"github.com/koding/kite/protocol"
)

const (
	Magic = ","
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	flag.Parse()

	k := util.NewKite("glenda", "1.0.0")

	k.HandleFunc("irc.privmsg", privmsgHandler).DisableAuthentication()

	var redial func()

	redial = func() {
		irc, err := util.IrcDial(k)
		if err != nil {
			k.Log.Fatal("Failed to connect to irc: %v", err)
		}

		err = util.IrcRegister(irc)
		if err != nil {
			k.Log.Fatal("Failed to register with irc: %v", err)
		}

		k.Log.Info("Registered to irc with URL: %v", irc.URL)
		irc.OnDisconnect(redial)
	}

	redial()

	<-k.ServerCloseNotify()
}

type Alias struct {
	KiteName   string
	MethodName string
	Arguments  []string
}

var (
	cmdAliases = map[string]Alias{
		"theo": Alias{"qdb", "qdb.getrandom", []string{"theo"}},
		"rob":  Alias{"qdb", "qdb.getrandom", []string{"rob"}},
		"ken":  Alias{"qdb", "qdb.getrandom", []string{"ken"}},
		"rsc":  Alias{"qdb", "qdb.getrandom", []string{"rsc"}},

		"troll": Alias{"qdb", "qdb.getrandom", []string{"troll"}},
		"terry": Alias{"qdb", "qdb.getrandom", []string{"terry"}},
		"roa":   Alias{"qdb", "qdb.getrandom", []string{"roa"}},
	}
)

func privmsgHandler(r *kite.Request) (result interface{}, err error) {
	kargs := map[string]string{}
	r.Args.One().MustUnmarshal(&kargs)

	r.LocalKite.Log.Debug("got a privmsg: %+v", kargs)

	msg := kargs["message"]
	args := strings.Split(msg, " ")

	if len(args) < 1 {
		return nil, nil
	}

	cmd := args[0]

	if !strings.HasPrefix(cmd, Magic) {
		return
	}

	cmd = strings.TrimPrefix(cmd, Magic)
	args = args[1:]

	r.LocalKite.Log.Debug("trying to call %q", cmd)

	cmdspl := strings.Split(cmd, ".")
	kitename := cmdspl[0]

	if alias, ok := cmdAliases[cmd]; ok {
		kitename = alias.KiteName
		cmd = alias.MethodName
		args = alias.Arguments
	} else {
		// "foo" invokes the method "foo.foo" on kite "foo"
		if len(cmdspl) == 1 {
			cmd = kitename + "." + kitename
		}
	}

	resp, err := call(r.LocalKite, kitename, cmd, args)
	if err != nil {
		kargs["message"] = fmt.Sprintf("calling %q failed: %v", cmd, err)
		_, err = r.Client.Tell("irc.privmsg", kargs)
		return nil, err
	}

	// kargs["reply-to"] is preserved
	kargs["message"] = resp

	_, err = r.Client.Tell("irc.privmsg", kargs)
	return nil, err
}

func call(k *kite.Kite, kitename, method string, args []string) (string, error) {
	kites, err := k.GetKites(&protocol.KontrolQuery{
		Username:    k.Config.Username,
		Environment: k.Config.Environment,
		Name:        kitename,
	})

	if err != nil {
		k.Log.Warning("call kite query error: %v", err)
		return "", fmt.Errorf("module doesn't exist")
	}

	var srv *kite.Client

	// take first working kite
	for _, kite := range kites {
		if err = kite.DialTimeout(10 * time.Second); err != nil {
			continue
		}

		srv = kite
		break
	}

	if err != nil {
		k.Log.Warning("call dial kite error: %v", err)
		return "", err
	}

	res, err := srv.Tell(method, args)
	if err != nil {
		k.Log.Warning("call invoke %q error: %v", method, err)
		return "", err
	}

	return res.MustString(), nil
}
