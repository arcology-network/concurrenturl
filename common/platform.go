package common

import (
	"strings"

	common "github.com/arcology-network/common-lib/common"
	commutative "github.com/arcology-network/concurrenturl/commutative"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
)

type Platform struct {
	syspaths map[string]uint8
}

func NewPlatform() *Platform {
	return &Platform{
		map[string]uint8{
			"/":                   commutative.PATH,
			"/code":               noncommutative.BYTES,
			"/nonce":              commutative.UINT64,
			"/balance":            commutative.UINT256,
			"/storage/":           commutative.PATH,
			"/storage/container/": commutative.PATH,
			"/storage/native/":    commutative.PATH,
			// "/storage/native/local/": commutative.PATH,
		},
	}
}

func Eth10AccountShard(numOfShard int, key string) int {
	if len(key) < 24 {
		panic("Invalid eth1.0 account shard key: " + key)
	}
	return (hex2int(key[22])*16 + hex2int(key[23])) % numOfShard
}

// func (this *Platform) RootLength() int { return len(this.Eth10Account()) + 40 }

func hex2int(c byte) int {
	if c >= 'a' {
		return int(c-'a') + 10
	} else {
		return int(c - '0')
	}
}

// Get ths builtin paths
func (this *Platform) GetBuiltins(acct string) ([]string, []uint8) {
	paths, typeIds := common.MapKVs(this.syspaths)
	common.SortBy1st(paths, typeIds, func(lhv, rhv string) bool { return lhv < rhv })

	for i, path := range paths {
		paths[i] = ETH10_ACCOUNT_PREFIX + acct + path
	}
	return paths, typeIds
}

// These paths won't keep the sub elements
func (this *Platform) IsSysPath(path string) bool {
	if len(path) <= ETH10_ACCOUNT_FULL_LENGTH {
		return path == ETH10 || path == ETH10_ACCOUNT_PREFIX
	}

	subPath := path[ETH10_ACCOUNT_FULL_LENGTH:] // Removed the shared part
	_, ok := this.syspaths[subPath]
	return ok
}

func (this *Platform) GetSysPaths() []string {
	return common.MapKeys(this.syspaths)
}

func (this *Platform) Builtins(acct string, idx int) string {
	paths, _ := common.MapKVs(this.syspaths)
	return ETH10_ACCOUNT_PREFIX + acct + paths[idx]
}

func ParseAccountAddr(acct string) (string, string, string) {
	if len(acct) < ETH10_ACCOUNT_PREFIX_LENGTH+ETH10_ACCOUNT_LENGTH {
		return acct, "", ""
	}
	return acct[:ETH10_ACCOUNT_PREFIX_LENGTH],
		acct[ETH10_ACCOUNT_PREFIX_LENGTH : ETH10_ACCOUNT_PREFIX_LENGTH+ETH10_ACCOUNT_LENGTH],
		acct[ETH10_ACCOUNT_PREFIX_LENGTH+ETH10_ACCOUNT_LENGTH:]
}

func UnderNative(key string) string {
	if len(key) >= ETH10_ACCOUNT_PREFIX_LENGTH+ETH10_ACCOUNT_LENGTH {
		subKey := key[ETH10_ACCOUNT_PREFIX_LENGTH+ETH10_ACCOUNT_LENGTH:]
		if strings.HasPrefix(subKey, "/storage/native/") {
			return key[len(subKey)+len("/storage/native/"):]
		}
	}
	return ""
}

func UnderContainer(key string) string {
	subKey := key[ETH10_ACCOUNT_PREFIX_LENGTH+ETH10_ACCOUNT_LENGTH:]
	if strings.HasPrefix(subKey, "/storage/container/") {
		return key[len(subKey)+len("/storage/container/"):]
	}
	return ""
}
