package rpc

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"strconv"
	"strings"
	"time"

	chain "github.com/phantasma-io/phantasma-sdk-go/pkg/blockchain"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/carbon"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/cryptography"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/jsonrpc"
	resp "github.com/phantasma-io/phantasma-sdk-go/pkg/rpc/response"
)

// PhantasmaRPC is a JSON-RPC client for Phantasma node endpoints.
type PhantasmaRPC struct {
	client jsonrpc.RPCClient
}

// AddressType selects how RPC should interpret address text.
type AddressType int

const (
	// AddressTypePhantasma treats account text as a Phantasma address.
	AddressTypePhantasma AddressType = iota
	// AddressTypeCarbon treats account text as a Carbon address.
	AddressTypeCarbon
)

// NewRPCMainnet returns an RPC client for the public mainnet endpoint.
func NewRPCMainnet() PhantasmaRPC {
	return NewRPC("https://pharpc1.phantasma.info/rpc")
}

// NewRPCSetMainnet returns RPC clients for the public mainnet endpoint set.
func NewRPCSetMainnet() []PhantasmaRPC {
	return []PhantasmaRPC{NewRPC("https://pharpc1.phantasma.info/rpc"),
		NewRPC("https://pharpc2.phantasma.info/rpc"),
		NewRPC("https://pharpc3.phantasma.info/rpc")}
}

// NewRPCTestnet returns an RPC client for the public testnet endpoint.
func NewRPCTestnet() PhantasmaRPC {
	return NewRPC("https://testnet.phantasma.info/rpc")
}

// NewRPC returns an RPC client for endpoint.
func NewRPC(endpoint string) PhantasmaRPC {
	rpc := PhantasmaRPC{
		client: jsonrpc.NewClient(endpoint),
	}
	return rpc
}

// Call performs a low-level JSON-RPC call with caller-supplied context.
func (rpc PhantasmaRPC) Call(ctx context.Context, method string, params ...interface{}) (*jsonrpc.RPCResponse, error) {
	return rpc.client.Call(normalizeContext(ctx), method, params...)
}

func normalizeContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func checkError(result *jsonrpc.RPCResponse, err error) error {
	if err != nil {
		return err
	}

	if result == nil {
		return errors.New("rpc response missing")
	}

	if result.Error != nil {
		return errors.New(result.Error.Message)
	}

	return nil
}

func (rpc PhantasmaRPC) callObject(ctx context.Context, out interface{}, method string, params ...interface{}) error {
	result, err := rpc.client.Call(normalizeContext(ctx), method, params...)
	if err := checkError(result, err); err != nil {
		return err
	}

	return result.GetObject(out)
}

func (rpc PhantasmaRPC) callString(ctx context.Context, method string, params ...interface{}) (string, error) {
	result, err := rpc.client.Call(normalizeContext(ctx), method, params...)
	if err := checkError(result, err); err != nil {
		return "", err
	}

	return result.GetString()
}

func (rpc PhantasmaRPC) callBool(ctx context.Context, method string, params ...interface{}) (bool, error) {
	result, err := rpc.client.Call(normalizeContext(ctx), method, params...)
	if err := checkError(result, err); err != nil {
		return false, err
	}

	if value, err := result.GetBool(); err == nil {
		return value, nil
	}

	var value bool
	if err := result.GetObject(&value); err != nil {
		var text string
		if textErr := result.GetObject(&text); textErr != nil {
			return false, err
		}
		return strconv.ParseBool(text)
	}
	return value, nil
}

func (rpc PhantasmaRPC) callInt(ctx context.Context, method string, params ...interface{}) (int, error) {
	result, err := rpc.client.Call(normalizeContext(ctx), method, params...)
	if err := checkError(result, err); err != nil {
		return 0, err
	}

	switch value := result.Result.(type) {
	case int:
		return value, nil
	case int64:
		return int(value), nil
	case float64:
		return int(value), nil
	case json.Number:
		parsed, err := value.Int64()
		return int(parsed), err
	case string:
		return strconv.Atoi(value)
	}

	var value int
	if err := result.GetObject(&value); err != nil {
		return 0, err
	}
	return value, nil
}

// GetPlatforms returns the platforms known by the connected node.
func (rpc PhantasmaRPC) GetPlatforms(ctx context.Context) ([]resp.PlatformResult, error) {
	var platforms []resp.PlatformResult
	result, err := rpc.client.Call(normalizeContext(ctx), "getPlatforms", []interface{}{})

	if err := checkError(result, err); err != nil {
		return []resp.PlatformResult{}, err
	}

	err = result.GetObject(&platforms)
	if err != nil {
		return []resp.PlatformResult{}, err
	}

	return platforms, nil
}

// GetAccounts returns account records for one or more addresses.
func (rpc PhantasmaRPC) GetAccounts(ctx context.Context, addresses ...string) ([]resp.AccountResult, error) {
	return rpc.GetAccountsText(ctx, strings.Join(addresses, ","))
}

// GetAccountsText returns account records for a comma-separated address list.
func (rpc PhantasmaRPC) GetAccountsText(ctx context.Context, addresses string) ([]resp.AccountResult, error) {
	var accounts []resp.AccountResult
	result, err := rpc.client.Call(normalizeContext(ctx), "getAccounts", addresses, false)

	if err := checkError(result, err); err != nil {
		return []resp.AccountResult{}, err
	}

	err = result.GetObject(&accounts)
	if err != nil {
		return []resp.AccountResult{}, err
	}

	return accounts, nil
}

// GetAccountsWithAddressType returns account records for addresses of the same address type.
func (rpc PhantasmaRPC) GetAccountsWithAddressType(ctx context.Context, addresses []string, extended bool, checkAddressReservedByte bool, addressType AddressType) ([]resp.AccountResult, error) {
	var accounts []resp.AccountResult
	if err := rpc.callObject(ctx, &accounts, "getAccounts", strings.Join(addresses, ","), extended, checkAddressReservedByte, addressType); err != nil {
		return []resp.AccountResult{}, err
	}
	return accounts, nil
}

// LookupName resolves a registered name into an address.
func (rpc PhantasmaRPC) LookupName(ctx context.Context, name string) (string, error) {
	result, err := rpc.client.Call(normalizeContext(ctx), "lookUpName", name)

	if err := checkError(result, err); err != nil {
		return "", err
	}

	address, err := result.GetString()
	if err != nil {
		return "", err
	}

	return address, nil
}

// GetAccount returns the current state of an account without the full transaction list.
func (rpc PhantasmaRPC) GetAccount(ctx context.Context, address string) (resp.AccountResult, error) {
	var account resp.AccountResult
	result, err := rpc.client.Call(normalizeContext(ctx), "getAccount", address, false)

	if err := checkError(result, err); err != nil {
		return resp.AccountResult{}, err
	}

	err = result.GetObject(&account)
	if err != nil {
		return resp.AccountResult{}, err
	}

	return account, nil
}

// GetAccountWithAddressType returns account state for an address with explicit address interpretation.
func (rpc PhantasmaRPC) GetAccountWithAddressType(ctx context.Context, address string, extended bool, checkAddressReservedByte bool, addressType AddressType) (resp.AccountResult, error) {
	var account resp.AccountResult
	if err := rpc.callObject(ctx, &account, "getAccount", address, extended, checkAddressReservedByte, addressType); err != nil {
		return resp.AccountResult{}, err
	}
	return account, nil
}

// GetAddressTransactions returns transactions for an address, ordered from newer to older.
func (rpc PhantasmaRPC) GetAddressTransactions(ctx context.Context, address string, page int, pageSize int) (resp.PaginatedResult[resp.AddressTransactionsResult], error) {
	var addressTxs resp.PaginatedResult[resp.AddressTransactionsResult]
	result, err := rpc.client.Call(normalizeContext(ctx), "getAddressTransactions", address, page, pageSize)

	if err := checkError(result, err); err != nil {
		return resp.PaginatedResult[resp.AddressTransactionsResult]{}, err
	}

	err = result.GetObject(&addressTxs)
	if err != nil {
		return resp.PaginatedResult[resp.AddressTransactionsResult]{}, err
	}

	return addressTxs, nil
}

// GetAddressTransactionCount returns the number of transactions for an address.
func (rpc PhantasmaRPC) GetAddressTransactionCount(ctx context.Context, address string, chainName string) (int, error) {
	var count int
	result, err := rpc.client.Call(normalizeContext(ctx), "getAddressTransactionCount", address, chainName)

	if err := checkError(result, err); err != nil {
		return 0, err
	}

	err = result.GetObject(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetBlockByHeight returns a block by chain and height.
func (rpc PhantasmaRPC) GetBlockByHeight(ctx context.Context, chain string, height string) (resp.BlockResult, error) {
	var blockResult resp.BlockResult
	result, err := rpc.client.Call(normalizeContext(ctx), "getBlockByHeight", chain, height)

	if err := checkError(result, err); err != nil {
		return resp.BlockResult{}, err
	}

	err = result.GetObject(&blockResult)
	if err != nil {
		errorResult := resp.ErrorResult{}
		err = result.GetObject(&errorResult)
		if err != nil {
			return blockResult, err
		}

		return blockResult, errors.New(errorResult.Error)
	}
	return blockResult, nil
}

// GetBlockHeight Returns height of the latest block minted on the chain
func (rpc PhantasmaRPC) GetBlockHeight(ctx context.Context, chainName string) (*big.Int, error) {
	var resultValue string
	result, err := rpc.client.Call(normalizeContext(ctx), "getBlockHeight", chainName)

	if err := checkError(result, err); err != nil {
		return big.NewInt(0), err
	}

	err = result.GetObject(&resultValue)
	if err != nil {
		return big.NewInt(0), err
	}

	height, ok := big.NewInt(0).SetString(resultValue, 10)
	if !ok {
		return big.NewInt(0), errors.New("invalid block height: " + resultValue)
	}
	return height, nil
}

// GetBlockTransactionCountByHash returns transaction count for a main-chain block hash.
func (rpc PhantasmaRPC) GetBlockTransactionCountByHash(ctx context.Context, blockHash string) (int, error) {
	return rpc.callInt(ctx, "getBlockTransactionCountByHash", "main", blockHash)
}

// GetBlockTransactionCountByHashOnChain returns transaction count for a block hash on a specific chain.
func (rpc PhantasmaRPC) GetBlockTransactionCountByHashOnChain(ctx context.Context, chainAddressOrName string, blockHash string) (int, error) {
	return rpc.callInt(ctx, "getBlockTransactionCountByHash", chainAddressOrName, blockHash)
}

// GetBlockByHash returns a block by hash.
func (rpc PhantasmaRPC) GetBlockByHash(ctx context.Context, blockHash string) (resp.BlockResult, error) {
	var blockResult resp.BlockResult
	if err := rpc.callObject(ctx, &blockResult, "getBlockByHash", blockHash); err != nil {
		return resp.BlockResult{}, err
	}
	return blockResult, nil
}

// GetLatestBlock returns the latest block for the given chain.
func (rpc PhantasmaRPC) GetLatestBlock(ctx context.Context, chainAddressOrName string) (resp.BlockResult, error) {
	var blockResult resp.BlockResult
	if err := rpc.callObject(ctx, &blockResult, "getLatestBlock", chainAddressOrName); err != nil {
		return resp.BlockResult{}, err
	}
	return blockResult, nil
}

// GetTransactionByBlockHashAndIndex returns a main-chain transaction by block hash and transaction index.
func (rpc PhantasmaRPC) GetTransactionByBlockHashAndIndex(ctx context.Context, blockHash string, index int) (resp.TransactionResult, error) {
	return rpc.GetTransactionByBlockHashAndIndexOnChain(ctx, "main", blockHash, index)
}

// GetTransactionByBlockHashAndIndexOnChain returns a transaction by chain, block hash and transaction index.
func (rpc PhantasmaRPC) GetTransactionByBlockHashAndIndexOnChain(ctx context.Context, chainAddressOrName string, blockHash string, index int) (resp.TransactionResult, error) {
	var txResult resp.TransactionResult
	if err := rpc.callObject(ctx, &txResult, "getTransactionByBlockHashAndIndex", chainAddressOrName, blockHash, index); err != nil {
		return resp.TransactionResult{}, err
	}
	return txResult, nil
}

// GetContract returns a contract by name on a chain.
func (rpc PhantasmaRPC) GetContract(ctx context.Context, name, chainName string) (resp.ContractResult, error) {
	var contract resp.ContractResult
	result, err := rpc.client.Call(normalizeContext(ctx), "getContract", chainName, name)

	if err := checkError(result, err); err != nil {
		return resp.ContractResult{}, err
	}

	err = result.GetObject(&contract)
	if err != nil {
		return resp.ContractResult{}, err
	}

	return contract, nil
}

// GetChains returns available chains.
func (rpc PhantasmaRPC) GetChains(ctx context.Context, extended bool) ([]resp.ChainResult, error) {
	var chains []resp.ChainResult
	if err := rpc.callObject(ctx, &chains, "getChains", extended); err != nil {
		return []resp.ChainResult{}, err
	}
	return chains, nil
}

// GetChain returns a chain by name.
func (rpc PhantasmaRPC) GetChain(ctx context.Context, name string, extended bool) (resp.ChainResult, error) {
	var chain resp.ChainResult
	if err := rpc.callObject(ctx, &chain, "getChain", name, extended); err != nil {
		return resp.ChainResult{}, err
	}
	return chain, nil
}

// GetNexus returns nexus metadata.
func (rpc PhantasmaRPC) GetNexus(ctx context.Context, extended bool) (resp.NexusResult, error) {
	var nexus resp.NexusResult
	if err := rpc.callObject(ctx, &nexus, "getNexus", extended); err != nil {
		return resp.NexusResult{}, err
	}
	return nexus, nil
}

// GetContracts returns contracts deployed on a chain.
func (rpc PhantasmaRPC) GetContracts(ctx context.Context, chainAddressOrName string, extended bool) ([]resp.ContractResult, error) {
	var contracts []resp.ContractResult
	if err := rpc.callObject(ctx, &contracts, "getContracts", chainAddressOrName, extended); err != nil {
		return []resp.ContractResult{}, err
	}
	return contracts, nil
}

// GetContractByName matches the RPC parameter order: chain, then contract name.
func (rpc PhantasmaRPC) GetContractByName(ctx context.Context, chainAddressOrName string, contractName string) (resp.ContractResult, error) {
	var contract resp.ContractResult
	if err := rpc.callObject(ctx, &contract, "getContract", chainAddressOrName, contractName); err != nil {
		return resp.ContractResult{}, err
	}
	return contract, nil
}

// GetContractByAddress returns a contract by its address.
func (rpc PhantasmaRPC) GetContractByAddress(ctx context.Context, chainAddressOrName string, contractAddress string) (resp.ContractResult, error) {
	var contract resp.ContractResult
	if err := rpc.callObject(ctx, &contract, "getContractByAddress", chainAddressOrName, contractAddress); err != nil {
		return resp.ContractResult{}, err
	}
	return contract, nil
}

// GetOrganization returns organization metadata by id.
func (rpc PhantasmaRPC) GetOrganization(ctx context.Context, id string, extended bool) (resp.OrganizationResult, error) {
	var organization resp.OrganizationResult
	if err := rpc.callObject(ctx, &organization, "getOrganization", id, extended); err != nil {
		return resp.OrganizationResult{}, err
	}
	return organization, nil
}

// GetOrganizationByName returns organization metadata by registered name.
func (rpc PhantasmaRPC) GetOrganizationByName(ctx context.Context, name string, extended bool) (resp.OrganizationResult, error) {
	var organization resp.OrganizationResult
	if err := rpc.callObject(ctx, &organization, "getOrganizationByName", name, extended); err != nil {
		return resp.OrganizationResult{}, err
	}
	return organization, nil
}

// GetOrganizations returns all organizations.
func (rpc PhantasmaRPC) GetOrganizations(ctx context.Context, extended bool) ([]resp.OrganizationResult, error) {
	var organizations []resp.OrganizationResult
	if err := rpc.callObject(ctx, &organizations, "getOrganizations", extended); err != nil {
		return []resp.OrganizationResult{}, err
	}
	return organizations, nil
}

// GetLeaderboard returns leaderboard rows by name.
func (rpc PhantasmaRPC) GetLeaderboard(ctx context.Context, name string) (resp.LeaderboardResult, error) {
	var leaderboard resp.LeaderboardResult
	if err := rpc.callObject(ctx, &leaderboard, "getLeaderboard", name); err != nil {
		return resp.LeaderboardResult{}, err
	}
	return leaderboard, nil
}

// InvokeRawScript executes a hex-encoded VM script against a chain without broadcasting it.
func (rpc PhantasmaRPC) InvokeRawScript(ctx context.Context, chain, script string) (resp.ScriptResult, error) {
	scriptResult := resp.ScriptResult{}
	result, err := rpc.client.Call(normalizeContext(ctx), "invokeRawScript", chain, script)

	if err := checkError(result, err); err != nil {
		return resp.ScriptResult{}, err
	}

	err = result.GetObject(&scriptResult)
	if err != nil {
		errorResult := resp.ErrorResult{}
		err = result.GetObject(&errorResult)
		if err != nil {
			return scriptResult, err
		}

		return scriptResult, errors.New(errorResult.Error)
	}

	return scriptResult, nil
}

// SendRawTransaction broadcasts a hex-encoded classic VM transaction.
func (rpc PhantasmaRPC) SendRawTransaction(ctx context.Context, txData string) (string, error) {
	var hash string
	result, err := rpc.client.Call(normalizeContext(ctx), "sendRawTransaction", txData)

	if err := checkError(result, err); err != nil {
		return "", err
	}

	hash, err = result.GetString()
	if err != nil {
		errorResult := resp.ErrorResult{}
		err = result.GetObject(&errorResult)
		if err != nil {
			return hash, err
		}

		return hash, errors.New(errorResult.Error)
	}

	return hash, nil
}

// SendCarbonTransaction broadcasts a serialized Carbon transaction.
func (rpc PhantasmaRPC) SendCarbonTransaction(ctx context.Context, txData string) (string, error) {
	return rpc.callString(ctx, "sendCarbonTransaction", txData)
}

// SignAndSendCarbonTransaction signs, serializes and broadcasts a Carbon transaction.
func (rpc PhantasmaRPC) SignAndSendCarbonTransaction(ctx context.Context, msg carbon.TxMsg, keys cryptography.KeyPair) (string, error) {
	signedTx, err := carbon.SignAndSerializeTxMsg(msg, keys)
	if err != nil {
		return "", err
	}
	return rpc.SendCarbonTransaction(ctx, hex.EncodeToString(signedTx))
}

// SignAndSendTransaction builds, signs and broadcasts a classic VM transaction with a binary payload.
func (rpc PhantasmaRPC) SignAndSendTransaction(ctx context.Context, keys cryptography.KeyPair, nexus string, script []byte, chainName string, payload []byte) (string, error) {
	return rpc.SignAndSendTransactionWithExpiration(ctx, keys, nexus, script, chainName, payload, uint32(time.Now().UTC().Add(20*time.Minute).Unix()))
}

// SignAndSendTransactionWithExpiration builds, signs and broadcasts a classic VM transaction with an explicit expiration.
func (rpc PhantasmaRPC) SignAndSendTransactionWithExpiration(ctx context.Context, keys cryptography.KeyPair, nexus string, script []byte, chainName string, payload []byte, expiration uint32) (string, error) {
	tx := chain.NewTransaction(nexus, chainName, script, expiration, payload)
	return rpc.SignAndSendBuiltTransaction(ctx, tx, keys)
}

// SignAndSendTransactionTextPayload builds, signs and broadcasts a classic VM transaction with a UTF-8 payload.
func (rpc PhantasmaRPC) SignAndSendTransactionTextPayload(ctx context.Context, keys cryptography.KeyPair, nexus string, script []byte, chainName string, payload string) (string, error) {
	return rpc.SignAndSendTransaction(ctx, keys, nexus, script, chainName, []byte(payload))
}

// SignAndSendBuiltTransaction signs and broadcasts an already-built classic VM transaction.
func (rpc PhantasmaRPC) SignAndSendBuiltTransaction(ctx context.Context, tx chain.Transaction, keys cryptography.KeyPair) (string, error) {
	if keys == nil {
		return "", errors.New("key pair is required")
	}
	if err := tx.Sign(keys); err != nil {
		return "", err
	}
	return rpc.SendRawTransaction(ctx, hex.EncodeToString(tx.Bytes()))
}

// GetTransaction returns transaction details by hash.
func (rpc PhantasmaRPC) GetTransaction(ctx context.Context, txHash string) (resp.TransactionResult, error) {
	var txResult resp.TransactionResult
	result, err := rpc.client.Call(normalizeContext(ctx), "getTransaction", txHash)

	if err := checkError(result, err); err != nil {
		return resp.TransactionResult{}, err
	}

	err = result.GetObject(&txResult)
	if err != nil {
		errorResult := resp.ErrorResult{}
		err = result.GetObject(&errorResult)
		if err != nil {
			return txResult, err
		}

		return txResult, errors.New(errorResult.Error)
	}
	return txResult, nil
}

// GetTokens returns token definitions known by the connected node.
func (rpc PhantasmaRPC) GetTokens(ctx context.Context, extended bool) ([]resp.TokenResult, error) {
	var txResult []resp.TokenResult
	result, err := rpc.client.Call(normalizeContext(ctx), "getTokens", extended)

	if err := checkError(result, err); err != nil {
		return []resp.TokenResult{}, err
	}

	err = result.GetObject(&txResult)
	if err != nil {
		errorResult := resp.ErrorResult{}
		err = result.GetObject(&errorResult)
		if err != nil {
			return txResult, err
		}

		return txResult, errors.New(errorResult.Error)
	}
	return txResult, nil
}

// GetTokensByOwner returns tokens filtered by owner address.
func (rpc PhantasmaRPC) GetTokensByOwner(ctx context.Context, extended bool, ownerAddress string) ([]resp.TokenResult, error) {
	var tokens []resp.TokenResult
	if err := rpc.callObject(ctx, &tokens, "getTokens", extended, ownerAddress); err != nil {
		return []resp.TokenResult{}, err
	}
	return tokens, nil
}

// GetTokensByOwnerWithAddressType returns tokens filtered by owner address and address type.
func (rpc PhantasmaRPC) GetTokensByOwnerWithAddressType(ctx context.Context, extended bool, ownerAddress string, addressType AddressType) ([]resp.TokenResult, error) {
	var tokens []resp.TokenResult
	if err := rpc.callObject(ctx, &tokens, "getTokens", extended, ownerAddress, addressType); err != nil {
		return []resp.TokenResult{}, err
	}
	return tokens, nil
}

// GetTokensAsMap returns chain tokens map where token symbol is used as a key
func (rpc PhantasmaRPC) GetTokensAsMap(ctx context.Context, extended bool) (map[string]resp.TokenResult, error) {
	result, err := rpc.GetTokens(ctx, extended)
	if err != nil {
		return nil, err
	}

	tokensMap := map[string]resp.TokenResult{}

	for _, t := range result {
		tokensMap[t.Symbol] = t
	}

	return tokensMap, nil
}

// GetToken returns a token by symbol.
func (rpc PhantasmaRPC) GetToken(ctx context.Context, symbol string, extended bool) (resp.TokenResult, error) {
	var txResult resp.TokenResult
	result, err := rpc.client.Call(normalizeContext(ctx), "getToken", symbol, extended)

	if err := checkError(result, err); err != nil {
		return resp.TokenResult{}, err
	}

	err = result.GetObject(&txResult)
	if err != nil {
		errorResult := resp.ErrorResult{}
		err = result.GetObject(&errorResult)
		if err != nil {
			return txResult, err
		}

		return txResult, errors.New(errorResult.Error)
	}
	return txResult, nil
}

// GetTokenWithID returns a token by symbol and optional Carbon token id. Use 0 when no Carbon id filter is needed.
func (rpc PhantasmaRPC) GetTokenWithID(ctx context.Context, symbol string, extended bool, carbonTokenID uint64) (resp.TokenResult, error) {
	var token resp.TokenResult
	if err := rpc.callObject(ctx, &token, "getToken", symbol, extended, carbonTokenID); err != nil {
		return resp.TokenResult{}, err
	}
	return token, nil
}

// GetTokenData returns token data for a Phantasma NFT id.
func (rpc PhantasmaRPC) GetTokenData(ctx context.Context, symbol string, tokenID string) (resp.TokenDataResult, error) {
	var tokenData resp.TokenDataResult
	if err := rpc.callObject(ctx, &tokenData, "getTokenData", symbol, tokenID); err != nil {
		return resp.TokenDataResult{}, err
	}
	return tokenData, nil
}

// GetTokenBalance returns a token balance for an account.
func (rpc PhantasmaRPC) GetTokenBalance(ctx context.Context, address string, tokenSymbol string, chainAddressOrName string) (resp.BalanceResult, error) {
	var balance resp.BalanceResult
	if err := rpc.callObject(ctx, &balance, "getTokenBalance", address, tokenSymbol, chainAddressOrName); err != nil {
		return resp.BalanceResult{}, err
	}
	return balance, nil
}

// GetTokenBalanceChecked returns a token balance and requests address reserved-byte validation.
func (rpc PhantasmaRPC) GetTokenBalanceChecked(ctx context.Context, address string, tokenSymbol string, chainAddressOrName string, checkAddressReservedByte bool) (resp.BalanceResult, error) {
	var balance resp.BalanceResult
	if err := rpc.callObject(ctx, &balance, "getTokenBalance", address, tokenSymbol, chainAddressOrName, checkAddressReservedByte); err != nil {
		return resp.BalanceResult{}, err
	}
	return balance, nil
}

// GetTokenBalanceWithAddressType returns a token balance with explicit address interpretation.
func (rpc PhantasmaRPC) GetTokenBalanceWithAddressType(ctx context.Context, address string, tokenSymbol string, chainAddressOrName string, checkAddressReservedByte bool, addressType AddressType) (resp.BalanceResult, error) {
	var balance resp.BalanceResult
	if err := rpc.callObject(ctx, &balance, "getTokenBalance", address, tokenSymbol, chainAddressOrName, checkAddressReservedByte, addressType); err != nil {
		return resp.BalanceResult{}, err
	}
	return balance, nil
}

// GetTokenSeries returns token series with cursor pagination. Use carbonTokenID 0 when no Carbon token id filter is needed.
func (rpc PhantasmaRPC) GetTokenSeries(ctx context.Context, symbol string, carbonTokenID uint64, pageSize int, cursor string) (resp.CursorPaginatedResult[[]resp.TokenSeriesResult], error) {
	var series resp.CursorPaginatedResult[[]resp.TokenSeriesResult]
	if err := rpc.callObject(ctx, &series, "getTokenSeries", symbol, carbonTokenID, pageSize, cursor); err != nil {
		return resp.CursorPaginatedResult[[]resp.TokenSeriesResult]{}, err
	}
	return series, nil
}

// GetTokenSeriesByID returns a single token series by Phantasma or Carbon series id.
func (rpc PhantasmaRPC) GetTokenSeriesByID(ctx context.Context, symbol string, carbonTokenID uint64, seriesID string, carbonSeriesID uint32) (resp.TokenSeriesResult, error) {
	var series resp.TokenSeriesResult
	if err := rpc.callObject(ctx, &series, "getTokenSeriesById", symbol, carbonTokenID, seriesID, carbonSeriesID); err != nil {
		return resp.TokenSeriesResult{}, err
	}
	return series, nil
}

// GetTokenNFTs returns NFTs for a Carbon token with cursor pagination.
func (rpc PhantasmaRPC) GetTokenNFTs(ctx context.Context, carbonTokenID uint64, carbonSeriesID uint32, pageSize int, cursor string, extended bool) (resp.CursorPaginatedResult[[]resp.TokenDataResult], error) {
	return rpc.GetTokenNFTsWithSeriesID(ctx, carbonTokenID, carbonSeriesID, "", pageSize, cursor, extended)
}

// GetTokenNFTsWithSeriesID returns NFTs and can filter by either Carbon or Phantasma series id.
func (rpc PhantasmaRPC) GetTokenNFTsWithSeriesID(ctx context.Context, carbonTokenID uint64, carbonSeriesID uint32, seriesID string, pageSize int, cursor string, extended bool) (resp.CursorPaginatedResult[[]resp.TokenDataResult], error) {
	var nfts resp.CursorPaginatedResult[[]resp.TokenDataResult]
	if err := rpc.callObject(ctx, &nfts, "getTokenNFTs", carbonTokenID, carbonSeriesID, pageSize, cursor, extended, seriesID); err != nil {
		return resp.CursorPaginatedResult[[]resp.TokenDataResult]{}, err
	}
	return nfts, nil
}

// GetAccountFungibleTokens returns fungible balances owned by an account with cursor pagination.
func (rpc PhantasmaRPC) GetAccountFungibleTokens(ctx context.Context, account string, tokenSymbol string, carbonTokenID uint64, pageSize int, cursor string, checkAddressReservedByte bool) (resp.CursorPaginatedResult[[]resp.BalanceResult], error) {
	var balances resp.CursorPaginatedResult[[]resp.BalanceResult]
	if err := rpc.callObject(ctx, &balances, "getAccountFungibleTokens", account, tokenSymbol, carbonTokenID, pageSize, cursor, checkAddressReservedByte); err != nil {
		return resp.CursorPaginatedResult[[]resp.BalanceResult]{}, err
	}
	return balances, nil
}

// GetAccountFungibleTokensWithAddressType returns fungible balances with explicit address interpretation.
func (rpc PhantasmaRPC) GetAccountFungibleTokensWithAddressType(ctx context.Context, account string, tokenSymbol string, carbonTokenID uint64, pageSize int, cursor string, checkAddressReservedByte bool, addressType AddressType) (resp.CursorPaginatedResult[[]resp.BalanceResult], error) {
	var balances resp.CursorPaginatedResult[[]resp.BalanceResult]
	if err := rpc.callObject(ctx, &balances, "getAccountFungibleTokens", account, tokenSymbol, carbonTokenID, pageSize, cursor, checkAddressReservedByte, addressType); err != nil {
		return resp.CursorPaginatedResult[[]resp.BalanceResult]{}, err
	}
	return balances, nil
}

// GetAccountNFTs returns NFTs owned by an account with cursor pagination.
func (rpc PhantasmaRPC) GetAccountNFTs(ctx context.Context, account string, tokenSymbol string, carbonTokenID uint64, carbonSeriesID uint32, pageSize int, cursor string, extended bool, checkAddressReservedByte bool) (resp.CursorPaginatedResult[[]resp.TokenDataResult], error) {
	var nfts resp.CursorPaginatedResult[[]resp.TokenDataResult]
	if err := rpc.callObject(ctx, &nfts, "getAccountNFTs", account, tokenSymbol, carbonTokenID, carbonSeriesID, pageSize, cursor, extended, checkAddressReservedByte); err != nil {
		return resp.CursorPaginatedResult[[]resp.TokenDataResult]{}, err
	}
	return nfts, nil
}

// GetAccountNFTsWithAddressType returns NFTs with explicit address interpretation.
func (rpc PhantasmaRPC) GetAccountNFTsWithAddressType(ctx context.Context, account string, tokenSymbol string, carbonTokenID uint64, carbonSeriesID uint32, pageSize int, cursor string, extended bool, checkAddressReservedByte bool, addressType AddressType) (resp.CursorPaginatedResult[[]resp.TokenDataResult], error) {
	var nfts resp.CursorPaginatedResult[[]resp.TokenDataResult]
	if err := rpc.callObject(ctx, &nfts, "getAccountNFTs", account, tokenSymbol, carbonTokenID, carbonSeriesID, pageSize, cursor, extended, checkAddressReservedByte, addressType); err != nil {
		return resp.CursorPaginatedResult[[]resp.TokenDataResult]{}, err
	}
	return nfts, nil
}

// GetAccountOwnedTokens returns token definitions owned by an account with cursor pagination.
func (rpc PhantasmaRPC) GetAccountOwnedTokens(ctx context.Context, account string, tokenSymbol string, carbonTokenID uint64, pageSize int, cursor string, checkAddressReservedByte bool) (resp.CursorPaginatedResult[[]resp.TokenResult], error) {
	var tokens resp.CursorPaginatedResult[[]resp.TokenResult]
	if err := rpc.callObject(ctx, &tokens, "getAccountOwnedTokens", account, tokenSymbol, carbonTokenID, pageSize, cursor, checkAddressReservedByte); err != nil {
		return resp.CursorPaginatedResult[[]resp.TokenResult]{}, err
	}
	return tokens, nil
}

// GetAccountOwnedTokensWithAddressType returns owned token definitions with explicit address interpretation.
func (rpc PhantasmaRPC) GetAccountOwnedTokensWithAddressType(ctx context.Context, account string, tokenSymbol string, carbonTokenID uint64, pageSize int, cursor string, checkAddressReservedByte bool, addressType AddressType) (resp.CursorPaginatedResult[[]resp.TokenResult], error) {
	var tokens resp.CursorPaginatedResult[[]resp.TokenResult]
	if err := rpc.callObject(ctx, &tokens, "getAccountOwnedTokens", account, tokenSymbol, carbonTokenID, pageSize, cursor, checkAddressReservedByte, addressType); err != nil {
		return resp.CursorPaginatedResult[[]resp.TokenResult]{}, err
	}
	return tokens, nil
}

// GetAccountOwnedTokenSeries returns token series owned by an account with cursor pagination.
func (rpc PhantasmaRPC) GetAccountOwnedTokenSeries(ctx context.Context, account string, tokenSymbol string, carbonTokenID uint64, pageSize int, cursor string, checkAddressReservedByte bool) (resp.CursorPaginatedResult[[]resp.TokenSeriesResult], error) {
	var series resp.CursorPaginatedResult[[]resp.TokenSeriesResult]
	if err := rpc.callObject(ctx, &series, "getAccountOwnedTokenSeries", account, tokenSymbol, carbonTokenID, pageSize, cursor, checkAddressReservedByte); err != nil {
		return resp.CursorPaginatedResult[[]resp.TokenSeriesResult]{}, err
	}
	return series, nil
}

// GetAccountOwnedTokenSeriesWithAddressType returns owned token series with explicit address interpretation.
func (rpc PhantasmaRPC) GetAccountOwnedTokenSeriesWithAddressType(ctx context.Context, account string, tokenSymbol string, carbonTokenID uint64, pageSize int, cursor string, checkAddressReservedByte bool, addressType AddressType) (resp.CursorPaginatedResult[[]resp.TokenSeriesResult], error) {
	var series resp.CursorPaginatedResult[[]resp.TokenSeriesResult]
	if err := rpc.callObject(ctx, &series, "getAccountOwnedTokenSeries", account, tokenSymbol, carbonTokenID, pageSize, cursor, checkAddressReservedByte, addressType); err != nil {
		return resp.CursorPaginatedResult[[]resp.TokenSeriesResult]{}, err
	}
	return series, nil
}

// GetAuctionsCount returns auction count for a token on a chain.
func (rpc PhantasmaRPC) GetAuctionsCount(ctx context.Context, chainAddressOrName string, symbol string) (int, error) {
	return rpc.callInt(ctx, "getAuctionsCount", chainAddressOrName, symbol)
}

// GetAuctions returns auctions with page pagination.
func (rpc PhantasmaRPC) GetAuctions(ctx context.Context, chainAddressOrName string, symbol string, page int, pageSize int) (resp.PaginatedResult[[]resp.AuctionResult], error) {
	var auctions resp.PaginatedResult[[]resp.AuctionResult]
	if err := rpc.callObject(ctx, &auctions, "getAuctions", chainAddressOrName, symbol, page, pageSize); err != nil {
		return resp.PaginatedResult[[]resp.AuctionResult]{}, err
	}
	return auctions, nil
}

// GetAuction returns a single auction by token id.
func (rpc PhantasmaRPC) GetAuction(ctx context.Context, chainAddressOrName string, symbol string, tokenID string) (resp.AuctionResult, error) {
	var auction resp.AuctionResult
	if err := rpc.callObject(ctx, &auction, "getAuction", chainAddressOrName, symbol, tokenID); err != nil {
		return resp.AuctionResult{}, err
	}
	return auction, nil
}

// GetNFT returns NFT data and optionally loads properties.
func (rpc PhantasmaRPC) GetNFT(ctx context.Context, symbol string, tokenID string, extended bool) (resp.TokenDataResult, error) {
	var nft resp.TokenDataResult
	if err := rpc.callObject(ctx, &nft, "getNFT", symbol, tokenID, extended); err != nil {
		return resp.TokenDataResult{}, err
	}
	return nft, nil
}

// GetNFTs returns NFT data for a list of token ids.
func (rpc PhantasmaRPC) GetNFTs(ctx context.Context, symbol string, tokenIDs []string, extended bool) ([]resp.TokenDataResult, error) {
	return rpc.GetNFTsText(ctx, symbol, strings.Join(tokenIDs, ","), extended)
}

// GetNFTsText returns NFT data for a comma-separated token id list.
func (rpc PhantasmaRPC) GetNFTsText(ctx context.Context, symbol string, tokenIDs string, extended bool) ([]resp.TokenDataResult, error) {
	var nfts []resp.TokenDataResult
	if err := rpc.callObject(ctx, &nfts, "getNFTs", symbol, tokenIDs, extended); err != nil {
		return []resp.TokenDataResult{}, err
	}
	return nfts, nil
}

// GetArchive returns archive metadata by hash.
func (rpc PhantasmaRPC) GetArchive(ctx context.Context, hash string) (resp.ArchiveResult, error) {
	var archive resp.ArchiveResult
	if err := rpc.callObject(ctx, &archive, "getArchive", hash); err != nil {
		return resp.ArchiveResult{}, err
	}
	return archive, nil
}

// WriteArchive writes an archive block.
func (rpc PhantasmaRPC) WriteArchive(ctx context.Context, hash string, blockIndex int, blockContent []byte) (bool, error) {
	return rpc.WriteArchiveBase64(ctx, hash, blockIndex, base64.StdEncoding.EncodeToString(blockContent))
}

// WriteArchiveBase64 writes a base64-encoded archive block.
func (rpc PhantasmaRPC) WriteArchiveBase64(ctx context.Context, hash string, blockIndex int, blockContent string) (bool, error) {
	return rpc.callBool(ctx, "writeArchive", hash, blockIndex, blockContent)
}

// ReadArchive reads a base64 encoded archive block.
func (rpc PhantasmaRPC) ReadArchive(ctx context.Context, hash string, blockIndex int) (string, error) {
	return rpc.callString(ctx, "readArchive", hash, blockIndex)
}

// GetVersion returns node build and version metadata.
func (rpc PhantasmaRPC) GetVersion(ctx context.Context) (resp.BuildInfoResult, error) {
	var buildInfo resp.BuildInfoResult
	result, err := rpc.client.Call(normalizeContext(ctx), "getVersion", []interface{}{})

	if err := checkError(result, err); err != nil {
		return resp.BuildInfoResult{}, err
	}

	err = result.GetObject(&buildInfo)
	if err != nil {
		return resp.BuildInfoResult{}, err
	}

	return buildInfo, nil
}

// GetPhantasmaVMConfig returns active VM configuration for a chain.
func (rpc PhantasmaRPC) GetPhantasmaVMConfig(ctx context.Context, chainAddressOrName string) (resp.PhantasmaVMConfigResult, error) {
	var config resp.PhantasmaVMConfigResult
	result, err := rpc.client.Call(normalizeContext(ctx), "getPhantasmaVmConfig", chainAddressOrName)

	if err := checkError(result, err); err != nil {
		return resp.PhantasmaVMConfigResult{}, err
	}

	err = result.GetObject(&config)
	if err != nil {
		return resp.PhantasmaVMConfigResult{}, err
	}

	return config, nil
}
