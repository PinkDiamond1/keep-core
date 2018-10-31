package gjkr

import "math/big"

// MemberCommitmentsMessage is a message payload that carries the sender's
// commitments to polynomial coefficients during distributed key generation.
//
// It is expected to be broadcast.
type MemberCommitmentsMessage struct {
	senderID int

	commitments []*big.Int // slice of `C_ik`
}

// PeerSharesMessage is a message payload that carries the sender's shares `s_ij`
// and `t_ij` for the recipient during distributed key generation.
//
// It is expected to be communicated in an encrypted fashion to the selected
// recipient.
type PeerSharesMessage struct {
	senderID   int
	receiverID int

	shareS *big.Int
	shareT *big.Int
}

// SecretSharesAccusationsMessage is a message payload that carries all of the
// sender's accusations against other members of the threshold group.
// If all other members behaved honestly from the sender's point of view, this
// message should be broadcast but with an empty slice of `accusedIDs`.
//
// It is expected to be broadcast.
type SecretSharesAccusationsMessage struct {
	senderID int

	accusedIDs []int
}
