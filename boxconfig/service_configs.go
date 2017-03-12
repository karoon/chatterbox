package boxconfig

import "chatterbox/mqtt/types"

type AuthConfiguration struct {
	Driver      types.AuthDriverType
	AuthNoMatch types.AuthActionType
	ACLNoMatch  types.AuthActionType
}
