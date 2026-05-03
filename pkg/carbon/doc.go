// Package carbon implements Phantasma Phoenix Carbon wire serialization, transaction
// messages, signing helpers, token-module argument/result blobs, and standard
// NFT metadata helpers used by the current Phantasma SDKs.
//
// Public builders that accept user-supplied metadata return errors. Functions
// prefixed with Must are intended for constants and tests where invalid input is
// a programmer error.
package carbon
