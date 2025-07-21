package config

// SOLANA_DEX_ADDRESS_TO_NAME 代币地址到名称的映射
var SOLANA_DEX_ADDRESS_TO_NAME = map[string]string{
	"11111111111111111111111111111111":             "SOL",
	"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v": "USDC",
	"Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB": "USDT",
	"So11111111111111111111111111111111111111112":  "WSOL",
	"mSoLzYCxHdYgdzU16g5QSh3i5K3z3KZK7ytfqcJm7So":  "mSOL",
	"7vfCXTUXx5WJV5JADk17DUJ4ksgau7utNKj4b963voxs": "ETH",
	"9n4nbM75f5Ui33ZbPYXn59EwSgE8CGsHtAeTH5YFeJ9E": "BTC",
	"SRMuApVNdxXokk5GT7XD5cUUgXMBCoAz2LHeuAoKWRt":  "SRM",
	"8wXtPeU6557ETkp9WHFY1n1EcU6NxDvbAggHGsMYiHsB": "GMGN",
}

// SOLANA_DEX_BASE_TOKEN 基础代币列表
var SOLANA_DEX_BASE_TOKEN = []string{
	"11111111111111111111111111111111",             // SOL
	"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
	"Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB", // USDT
	"So11111111111111111111111111111111111111112",  // WSOL
}

// SOLANA_DEX_STABLE_TOKEN 稳定币列表
var SOLANA_DEX_STABLE_TOKEN = []string{
	"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
	"Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB", // USDT
}

// BLACK_LIST_TOKEN 黑名单代币
var BLACK_LIST_TOKEN = []string{
	"11111111111111111111111111111111",             // SOL (某些情况下需要过滤)
	"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",  // Token Program
	"ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL", // Associated Token Program
}

// WALLET_BLACKLIST 钱包黑名单
var WALLET_BLACKLIST = []string{
	"11111111111111111111111111111111",            // System Program
	"Vote111111111111111111111111111111111111111", // Vote Program
	"Stake11111111111111111111111111111111111111", // Stake Program
}

// MEVBOT_ADDRESSES MEV 机器人地址
var MEVBOT_ADDRESSES = []string{
	"5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1", // Example MEV bot
	"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",  // Example MEV bot
}

// IsBaseToken 检查是否为基础代币
func IsBaseToken(tokenAddress string) bool {
	for _, baseToken := range SOLANA_DEX_BASE_TOKEN {
		if baseToken == tokenAddress {
			return true
		}
	}
	return false
}

// IsStableToken 检查是否为稳定币
func IsStableToken(tokenAddress string) bool {
	for _, stableToken := range SOLANA_DEX_STABLE_TOKEN {
		if stableToken == tokenAddress {
			return true
		}
	}
	return false
}

// IsBlacklistedToken 检查是否为黑名单代币
func IsBlacklistedToken(tokenAddress string) bool {
	for _, blackToken := range BLACK_LIST_TOKEN {
		if blackToken == tokenAddress {
			return true
		}
	}
	return false
}

// IsBlacklistedWallet 检查是否为黑名单钱包
func IsBlacklistedWallet(walletAddress string) bool {
	for _, blackWallet := range WALLET_BLACKLIST {
		if blackWallet == walletAddress {
			return true
		}
	}
	return false
}

// IsMEVBot 检查是否为 MEV 机器人
func IsMEVBot(walletAddress string) bool {
	for _, mevBot := range MEVBOT_ADDRESSES {
		if mevBot == walletAddress {
			return true
		}
	}
	return false
}
