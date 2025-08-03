package mqtt

import (
	"bytes"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

// AllowHook is an authentication hook which allows connection access
// for all users and read and write access to all topics.
type AllowHook struct {
	mqtt.HookBase
	auth *HookAuth
}

// ID returns the ID of the hook.
func (h *AllowHook) ID() string {
	return "allow-all-auth"
}

func (h *AllowHook) Init(config any) error {
	h.auth = config.(*HookAuth)
	return nil
}

// Provides indicates which hook methods this hook provides.
func (h *AllowHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnConnectAuthenticate,
		mqtt.OnACLCheck,
		mqtt.OnDisconnect,
	}, []byte{b})
}

// OnConnectAuthenticate returns true/allowed for all requests.
func (h *AllowHook) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {
	return h.auth.login(cl.ID, string(pk.Connect.Username), string(pk.Connect.Password))
}

// OnACLCheck returns true/allowed for all checks.
func (h *AllowHook) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	return h.auth.ACLCheck(cl.ID, topic, write)
}

func (h *AllowHook) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
	// código...
	h.auth.del(cl.ID)
}
