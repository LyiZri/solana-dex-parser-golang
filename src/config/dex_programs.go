package config

// DexProgram 定义 DEX 协议结构
type DexProgram struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// 定义所有支持的 DEX 协议程序 ID
var DEX_PROGRAMS = map[string]DexProgram{
	// Jupiter 聚合器系列
	"JUPITER": {
		ID:   "JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4",
		Name: "Jupiter",
	},
	"JUPITER_DCA": {
		ID:   "DCA265Vj8a9CEuX1eb1LWRnDT7uK6q1xMipnNyatn23M",
		Name: "Jupiter DCA",
	},
	"JUPITER_LIMIT_ORDER": {
		ID:   "jupoNjAxXgZ4rjzxzPMP4oxduvQsQtZzyknqvzYNrNu",
		Name: "Jupiter Limit Order",
	},
	"JUPITER_LIMIT_ORDER_V2": {
		ID:   "j1o2qRpjcyUwEvwtcfhEQefh773ZgjxcVRry7LDqg5X",
		Name: "Jupiter Limit Order V2",
	},
	"JUPITER_VA": {
		ID:   "JUPyiwrYJFskUPiHa7hkeR8VUtAeFoSYbKedZNsDvCN",
		Name: "Jupiter VA",
	},
	"JUPITER_DCA_KEEPER1": {
		ID:   "Cw8CFyM9FkoMi7K7Ci6Q6aaSuC9jFeXkNM4euhrkQktU",
		Name: "Jupiter DCA Keeper 1",
	},
	"JUPITER_DCA_KEEPER2": {
		ID:   "KeEpR1q8N8kC2HcjCVqVyFxcDPs2Gvr5bLZnJvMJWCF",
		Name: "Jupiter DCA Keeper 2",
	},
	"JUPITER_DCA_KEEPER3": {
		ID:   "KEp5R2Q8N8kC2HcjCVqVyFxcDPs2Gvr5bLZnJvMJWCF",
		Name: "Jupiter DCA Keeper 3",
	},

	// Raydium 系列
	"RAYDIUM_V4": {
		ID:   "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8",
		Name: "Raydium V4",
	},
	"RAYDIUM_ROUTE": {
		ID:   "routeUGWgWzqBWFcrCfv8tritsqukccJPu3q5GPP3xS",
		Name: "Raydium Route",
	},
	"RAYDIUM_CPMM": {
		ID:   "CPMMoo8L3F4NbTegBCKVNunggL7H1ZpdTHKxQB5qKP1C",
		Name: "Raydium CPMM",
	},
	"RAYDIUM_CL": {
		ID:   "CAMMCzo5YL8w4VFF8KVHrK22GGUQpMkjJSyQdqWxjtGw",
		Name: "Raydium Concentrated Liquidity",
	},
	"RAYDIUM_LCP": {
		ID:   "5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1",
		Name: "Raydium Launchpad",
	},

	// Orca 系列
	"ORCA": {
		ID:   "whirLbMiicVdio4qvUfM5KAg6Ct8VwpYzGff3uctyCc",
		Name: "Orca Whirlpool",
	},
	"ORCA_V1": {
		ID:   "9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP",
		Name: "Orca V1",
	},
	"ORCA_V2": {
		ID:   "DjVE6JNiYqPL2QXyCUUh8rNjHrbz9hXHNYt99MQ59qw1",
		Name: "Orca V2",
	},

	// Meteora 系列
	"METEORA": {
		ID:   "Eo7WjKq67rjJQSZxS6z3YkapzY3eMj6Xy8X5EQVn5UaB",
		Name: "Meteora DLMM",
	},
	"METEORA_POOLS": {
		ID:   "24Uqj9JCLxUeoC3hGfh5W3s9FM9uCHDS2SG3LYwBpyTi",
		Name: "Meteora Pools",
	},
	"METEORA_DAMM": {
		ID:   "AMM55ShdkoGRB5jVYPjWziwk8m5MpwyDgsMWHaMSQWH6",
		Name: "Meteora DAMM",
	},

	// Meme 代币平台
	"PUMP_FUN": {
		ID:   "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P",
		Name: "Pump.fun",
	},
	"PUMP_SWAP": {
		ID:   "6Ef8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P",
		Name: "Pumpswap",
	},
	"MOONSHOT": {
		ID:   "MoonCVVNZFSYkqNXP6bxHLPL6QQJiMagDL3qcqUQTrG",
		Name: "Moonshot",
	},
	"BOOP_FUN": {
		ID:   "Boop6LvwCp5xqWMpPQ9V6kF4LuvJzgQaGKjGG7RSSJa",
		Name: "Boop.fun",
	},

	// 交易机器人
	"BANANA_GUN": {
		ID:   "BANKkqyHZbGgZNjHc4bYqTAg5zLnHHxXFmH7X5RKYou",
		Name: "Banana Gun",
	},
	"MINTECH": {
		ID:   "MinTvx2Z8qZZ8Q7hZqHqQz1Hh7BgEKQ5ZP7W5oJ4tN",
		Name: "Mintech",
	},
	"BLOOM": {
		ID:   "BLM5rYLZWs1U9u5U5u5u5u5u5u5u5u5u5u5u5u5u5u",
		Name: "Bloom",
	},
	"MAESTRO": {
		ID:   "MAEStRo1Y8Z8qZZ8Q7hZqHqQz1Hh7BgEKQ5ZP7W5oJ4",
		Name: "Maestro",
	},
	"NOVA": {
		ID:   "NoVa5rYLZWs1U9u5U5u5u5u5u5u5u5u5u5u5u5u5u5u",
		Name: "Nova",
	},
	"APEPRO": {
		ID:   "ApEPRo1Y8Z8qZZ8Q7hZqHqQz1Hh7BgEKQ5ZP7W5oJ4",
		Name: "Apepro",
	},

	// 其他主要 DEX
	"PHOENIX": {
		ID:   "PhoeNiXZ8ByJGLkxNfZRnkUfjvmuYqLkk2YcQcwcvN",
		Name: "Phoenix",
	},
	"OPENBOOK": {
		ID:   "srmqPvymJeFKQ4zGQed1GFppgkRHL9kaELCbyksJtPX",
		Name: "Openbook",
	},
	"ALDRIN": {
		ID:   "AmmV4V8r6VYA7ZJm7kn5KY1LGhKFsKN7pJiF4CFHKaV",
		Name: "Aldrin",
	},
	"CREMA": {
		ID:   "CremaBjb7LYe9MmhZzwpJz3L9bZ5KJ5pJ4CFHKaV2z",
		Name: "Crema",
	},
	"GOOSEFX": {
		ID:   "GXj4pzqwAMhbGzveFjFKjfzKtSWHkTKdp5CpQvHj7ZYS",
		Name: "GooseFX",
	},
	"LIFINITY": {
		ID:   "EewxydAPCCVuNEyrVN68PuSYdQ7wKn27V9Gjeoi8dy3S",
		Name: "Lifinity",
	},
	"MERCURIAL": {
		ID:   "MERLuDFBMmsHnsBPZw2sDQZHvXFMwp8EdjudcU2HKky",
		Name: "Mercurial",
	},
	"SABER": {
		ID:   "SSwpkEEcbUqx4vtoEByFjSkhKdCT862DNVb52nZg1UZ",
		Name: "Saber",
	},
	"SAROS": {
		ID:   "SarosF1KBSqWH1VnJjBZAGfY2Lq2vDKfLKKnUKJZh1M",
		Name: "Saros",
	},
	"SOLFI": {
		ID:   "SoLFiAmVW3FAXCJqUbH3MV1Nw4kEQ5ZP7W5oJ4tN2Z",
		Name: "SolFi",
	},
	"STABBLE": {
		ID:   "StabBLEr1Y8Z8qZZ8Q7hZqHqQz1Hh7BgEKQ5ZP7W5",
		Name: "Stabble",
	},
	"SANCTUM": {
		ID:   "SanCtuMQdVEz1c7vdxQJVAZP7W5oJ4tN2Zq8Q7hZq",
		Name: "Sanctum",
	},
	"PHOTON": {
		ID:   "PhOtOnMQdVEz1c7vdxQJVAZP7W5oJ4tN2Zq8Q7hZq",
		Name: "Photon",
	},
	"OKX_DEX": {
		ID:   "OKXdExMQdVEz1c7vdxQJVAZP7W5oJ4tN2Zq8Q7hZq",
		Name: "OKX DEX",
	},
}

// GetDexProgramByID 根据程序 ID 获取 DEX 信息
func GetDexProgramByID(programID string) (DexProgram, bool) {
	for _, program := range DEX_PROGRAMS {
		if program.ID == programID {
			return program, true
		}
	}
	return DexProgram{}, false
}

// GetAllProgramIDs 获取所有程序 ID 列表
func GetAllProgramIDs() []string {
	var programIDs []string
	for _, program := range DEX_PROGRAMS {
		programIDs = append(programIDs, program.ID)
	}
	return programIDs
}

// IsJupiterFamily 判断是否为 Jupiter 系列协议
func IsJupiterFamily(programID string) bool {
	jupiterIDs := []string{
		DEX_PROGRAMS["JUPITER"].ID,
		DEX_PROGRAMS["JUPITER_DCA"].ID,
		DEX_PROGRAMS["JUPITER_LIMIT_ORDER"].ID,
		DEX_PROGRAMS["JUPITER_LIMIT_ORDER_V2"].ID,
		DEX_PROGRAMS["JUPITER_VA"].ID,
		DEX_PROGRAMS["JUPITER_DCA_KEEPER1"].ID,
		DEX_PROGRAMS["JUPITER_DCA_KEEPER2"].ID,
		DEX_PROGRAMS["JUPITER_DCA_KEEPER3"].ID,
	}

	for _, id := range jupiterIDs {
		if id == programID {
			return true
		}
	}
	return false
}

// GetProgramName 根据程序 ID 获取程序名称
func GetProgramName(programID string) string {
	if program, exists := GetDexProgramByID(programID); exists {
		return program.Name
	}
	return "Unknown"
}
