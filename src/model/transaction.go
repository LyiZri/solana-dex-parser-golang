package model

// RPC request structure for Solana JSON-RPC calls
type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type GetBlockResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      int       `json:"id"`
	Result  *Block    `json:"result"`
	Error   *RPCError `json:"error,omitempty"`
}

// RPC response structure
type RPCResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      int       `json:"id"`
	Result  *Block    `json:"result"`
	Error   *RPCError `json:"error,omitempty"`
}

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Transaction structure
type Transaction struct {
	Signatures []string           `json:"signatures"`
	Message    TransactionMessage `json:"message"`
	Meta       *TransactionMeta   `json:"meta,omitempty"`
}

type TransactionMessage struct {
	AccountKeys         []string                 `json:"accountKeys"`
	Header              TransactionHeader        `json:"header"`
	Instructions        []TransactionInstruction `json:"instructions"`
	RecentBlockhash     string                   `json:"recentBlockhash"`
	AddressTableLookups []AddressTableLookup     `json:"addressTableLookups,omitempty"`
}

type TransactionHeader struct {
	NumRequiredSignatures       int `json:"numRequiredSignatures"`
	NumReadonlySignedAccounts   int `json:"numReadonlySignedAccounts"`
	NumReadonlyUnsignedAccounts int `json:"numReadonlyUnsignedAccounts"`
}

type TransactionInstruction struct {
	ProgramIdIndex int    `json:"programIdIndex"`
	Accounts       []int  `json:"accounts"`
	Data           string `json:"data"`
}

type AddressTableLookup struct {
	AccountKey      string `json:"accountKey"`
	WritableIndexes []int  `json:"writableIndexes"`
	ReadonlyIndexes []int  `json:"readonlyIndexes"`
}

type TransactionMeta struct {
	Err                  interface{}            `json:"err"`
	Status               map[string]interface{} `json:"status"`
	Fee                  uint64                 `json:"fee"`
	PreBalances          []uint64               `json:"preBalances"`
	PostBalances         []uint64               `json:"postBalances"`
	InnerInstructions    []InnerInstruction     `json:"innerInstructions,omitempty"`
	LogMessages          []string               `json:"logMessages,omitempty"`
	PreTokenBalances     []TokenBalance         `json:"preTokenBalances,omitempty"`
	PostTokenBalances    []TokenBalance         `json:"postTokenBalances,omitempty"`
	Rewards              []Reward               `json:"rewards,omitempty"`
	LoadedAddresses      *LoadedAddresses       `json:"loadedAddresses,omitempty"`
	ComputeUnitsConsumed *uint64                `json:"computeUnitsConsumed,omitempty"`
}

type InnerInstruction struct {
	Index        int                      `json:"index"`
	Instructions []TransactionInstruction `json:"instructions"`
}

type TokenBalance struct {
	AccountIndex  int           `json:"accountIndex"`
	Mint          string        `json:"mint"`
	Owner         string        `json:"owner,omitempty"`
	ProgramId     string        `json:"programId,omitempty"`
	UiTokenAmount UiTokenAmount `json:"uiTokenAmount"`
}

type UiTokenAmount struct {
	Amount         string   `json:"amount"`
	Decimals       int      `json:"decimals"`
	UiAmount       *float64 `json:"uiAmount"`
	UiAmountString string   `json:"uiAmountString"`
}

type Reward struct {
	Pubkey      string `json:"pubkey"`
	Lamports    int64  `json:"lamports"`
	PostBalance uint64 `json:"postBalance"`
	RewardType  string `json:"rewardType,omitempty"`
	Commission  *int   `json:"commission,omitempty"`
}

type LoadedAddresses struct {
	Writable []string `json:"writable"`
	Readonly []string `json:"readonly"`
}

// Block structure representing a Solana block
type Block struct {
	BlockHeight       *uint64       `json:"blockHeight"`
	BlockTime         *int64        `json:"blockTime"`
	Blockhash         string        `json:"blockhash"`
	ParentSlot        uint64        `json:"parentSlot"`
	PreviousBlockhash string        `json:"previousBlockhash"`
	Transactions      []Transaction `json:"transactions"`
	Rewards           []Reward      `json:"rewards,omitempty"`
}

// BlockResult represents the result of a block fetch operation
type BlockResult struct {
	Block Block
	Slot  uint64
	Error error
}
