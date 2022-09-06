package dkg

import (
	"context"
	"fmt"
	"github.com/keep-network/keep-core/pkg/tecdsa/common"

	"github.com/bnb-chain/tss-lib/tss"
	"github.com/keep-network/keep-core/pkg/crypto/ephemeral"
	"github.com/keep-network/keep-core/pkg/protocol/group"
)

// generateEphemeralKeyPair takes the group member list and generates an
// ephemeral ECDH keypair for every other group member. Generated public
// ephemeral keys are broadcasted within the group.
func (ekpgm *ephemeralKeyPairGeneratingMember) generateEphemeralKeyPair() (
	*ephemeralPublicKeyMessage,
	error,
) {
	ephemeralKeys := make(map[group.MemberIndex]*ephemeral.PublicKey)

	// Calculate ephemeral key pair for every other group member
	for _, member := range ekpgm.group.MemberIDs() {
		if member == ekpgm.id {
			// don’t actually generate a key with ourselves
			continue
		}

		ephemeralKeyPair, err := ephemeral.GenerateKeyPair()
		if err != nil {
			return nil, err
		}

		// save the generated ephemeral key to our state
		ekpgm.ephemeralKeyPairs[member] = ephemeralKeyPair

		// store the public key to the map for the message
		ephemeralKeys[member] = ephemeralKeyPair.PublicKey
	}

	return &ephemeralPublicKeyMessage{
		senderID:            ekpgm.id,
		ephemeralPublicKeys: ephemeralKeys,
		sessionID:           ekpgm.sessionID,
	}, nil
}

// generateSymmetricKeys attempts to generate symmetric keys for all remote group
// members via ECDH. It generates this symmetric key for each remote group member
// by doing an ECDH between the ephemeral private key generated for a remote
// group member, and the public key for this member, generated and broadcasted by
// the remote group member.
func (skgm *symmetricKeyGeneratingMember) generateSymmetricKeys(
	ephemeralPubKeyMessages []*ephemeralPublicKeyMessage,
) error {
	for _, ephemeralPubKeyMessage := range deduplicateBySender(ephemeralPubKeyMessages) {
		otherMember := ephemeralPubKeyMessage.senderID

		if !skgm.isValidEphemeralPublicKeyMessage(ephemeralPubKeyMessage) {
			return fmt.Errorf(
				"member [%v] sent invalid ephemeral public key message",
				otherMember,
			)
		}

		// Find the ephemeral key pair generated by this group member for
		// the other group member.
		ephemeralKeyPair, ok := skgm.ephemeralKeyPairs[otherMember]
		if !ok {
			return fmt.Errorf(
				"ephemeral key pair does not exist for member [%v]",
				otherMember,
			)
		}

		// Get the ephemeral private key generated by this group member for
		// the other group member.
		thisMemberEphemeralPrivateKey := ephemeralKeyPair.PrivateKey

		// Get the ephemeral public key broadcasted by the other group member,
		// which was intended for this group member.
		otherMemberEphemeralPublicKey :=
			ephemeralPubKeyMessage.ephemeralPublicKeys[skgm.id]

		// Create symmetric key for the current group member and the other
		// group member by ECDH'ing the public and private key.
		symmetricKey := thisMemberEphemeralPrivateKey.Ecdh(
			otherMemberEphemeralPublicKey,
		)
		skgm.symmetricKeys[otherMember] = symmetricKey
	}

	return nil
}

// isValidEphemeralPublicKeyMessage validates a given EphemeralPublicKeyMessage.
// Message is considered valid if it contains ephemeral public keys for
// all other group members.
func (skgm *symmetricKeyGeneratingMember) isValidEphemeralPublicKeyMessage(
	message *ephemeralPublicKeyMessage,
) bool {
	for _, memberID := range skgm.group.MemberIDs() {
		if memberID == message.senderID {
			// Message contains ephemeral public keys only for other group members
			continue
		}

		if _, ok := message.ephemeralPublicKeys[memberID]; !ok {
			skgm.logger.Warningf(
				"[member:%v] ephemeral public key message from member [%v] "+
					"does not contain public key for member [%v]",
				skgm.id,
				message.senderID,
				memberID,
			)
			return false
		}
	}

	return true
}

// tssRoundOne starts the TSS process by executing its first round. The
// outcome of that round is a message containing commitments and Paillier
// public keys for all other group members.
func (trom *tssRoundOneMember) tssRoundOne(
	ctx context.Context,
) (*tssRoundOneMessage, error) {
	if err := trom.tssParty.Start(); err != nil {
		return nil, fmt.Errorf(
			"failed to start TSS round one: [%v]",
			err,
		)
	}

	// We expect exactly one TSS message to be produced in this phase.
	select {
	case tssMessage := <-trom.tssOutgoingMessagesChan:
		tssMessageBytes, _, err := tssMessage.WireBytes()
		if err != nil {
			return nil, fmt.Errorf(
				"failed to encode TSS round one message: [%v]",
				err,
			)
		}

		return &tssRoundOneMessage{
			senderID:  trom.id,
			payload:   tssMessageBytes,
			sessionID: trom.sessionID,
		}, nil
	case <-ctx.Done():
		return nil, fmt.Errorf(
			"TSS round one outgoing message was not generated on time",
		)
	}
}

// tssRoundTwo performs the second round of the TSS process. The outcome of
// that round is a message containing shares and de-commitments for all other
// group members.
func (trtm *tssRoundTwoMember) tssRoundTwo(
	ctx context.Context,
	tssRoundOneMessages []*tssRoundOneMessage,
) (*tssRoundTwoMessage, error) {
	// Use messages from round one to update the local party and advance
	// to round two.
	for _, tssRoundOneMessage := range deduplicateBySender(tssRoundOneMessages) {
		senderID := tssRoundOneMessage.SenderID()

		_, tssErr := trtm.tssParty.UpdateFromBytes(
			tssRoundOneMessage.payload,
			common.ResolveSortedTssPartyID(trtm.tssParameters, senderID),
			true,
		)
		if tssErr != nil {
			return nil, fmt.Errorf(
				"cannot update using TSS round one message "+
					"from member [%v]: [%v]",
				senderID,
				tssErr,
			)
		}
	}

	// Listen for TSS outgoing messages. We expect N-1 P2P messages (where N
	// is the number of properly operating members) carrying shares and 1
	// broadcast message holding the de-commitments.
	var tssMessages []tss.Message
outgoingMessagesLoop:
	for {
		select {
		case tssMessage := <-trtm.tssOutgoingMessagesChan:
			tssMessages = append(tssMessages, tssMessage)

			if len(tssMessages) == len(trtm.group.OperatingMemberIDs()) {
				break outgoingMessagesLoop
			}
		case <-ctx.Done():
			return nil, fmt.Errorf(
				"TSS round two outgoing messages were not " +
					"generated on time",
			)
		}
	}

	broadcastPayload, peersPayload, err := common.AggregateTssMessages(
		tssMessages,
		trtm.symmetricKeys,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot aggregate TSS round two outgoing messages: [%w]",
			err,
		)
	}

	ok := len(broadcastPayload) > 0 &&
		len(peersPayload) == len(trtm.group.OperatingMemberIDs())-1
	if !ok {
		return nil, fmt.Errorf("cannot produce a proper TSS round two message")
	}

	return &tssRoundTwoMessage{
		senderID:         trtm.id,
		broadcastPayload: broadcastPayload,
		peersPayload:     peersPayload,
		sessionID:        trtm.sessionID,
	}, nil
}

// tssRoundThree performs the third round of the TSS process. The outcome of
// that round is a message containing Paillier proofs for all other group
// members.
func (trtm *tssRoundTwoMember) tssRoundThree(
	ctx context.Context,
	tssRoundTwoMessages []*tssRoundTwoMessage,
) (*tssRoundThreeMessage, error) {
	// Use messages from round two to update the local party and advance
	// to round three.
	for _, tssRoundTwoMessage := range deduplicateBySender(tssRoundTwoMessages) {
		senderID := tssRoundTwoMessage.SenderID()
		senderTssPartyID := common.ResolveSortedTssPartyID(trtm.tssParameters, senderID)

		// Update the local TSS party using the broadcast part of the message
		// produced in round two.
		_, tssErr := trtm.tssParty.UpdateFromBytes(
			tssRoundTwoMessage.broadcastPayload,
			senderTssPartyID,
			true,
		)
		if tssErr != nil {
			return nil, fmt.Errorf(
				"cannot update using the broadcast part of the "+
					"TSS round two message from member [%v]: [%v]",
				senderID,
				tssErr,
			)
		}

		// Check if the sender produced a P2P part of the TSS round two message
		// for this member.
		encryptedPeerPayload, ok := tssRoundTwoMessage.peersPayload[trtm.id]
		if !ok {
			return nil, fmt.Errorf(
				"no P2P part in the TSS round two message from member [%v]",
				senderID,
			)
		}
		// Get the symmetric key with the sender. If the symmetric key
		// cannot be found, something awful happened.
		symmetricKey, ok := trtm.symmetricKeys[senderID]
		if !ok {
			return nil, fmt.Errorf(
				"cannot get symmetric key with member [%v]",
				senderID,
			)
		}
		// Decrypt the P2P part of the TSS round two message.
		peerPayload, err := symmetricKey.Decrypt(encryptedPeerPayload)
		if err != nil {
			return nil, fmt.Errorf(
				"cannot decrypt P2P part of the TSS round two "+
					"message from member [%v]: [%v]",
				senderID,
				err,
			)
		}
		// Update the local TSS party using the P2P part of the message
		// produced in round two.
		_, tssErr = trtm.tssParty.UpdateFromBytes(
			peerPayload,
			senderTssPartyID,
			false,
		)
		if tssErr != nil {
			return nil, fmt.Errorf(
				"cannot update using the P2P part of the TSS round "+
					"two message from member [%v]: [%v]",
				senderID,
				tssErr,
			)
		}
	}

	// We expect exactly one TSS message to be produced in this phase.
	select {
	case tssMessage := <-trtm.tssOutgoingMessagesChan:
		tssMessageBytes, _, err := tssMessage.WireBytes()
		if err != nil {
			return nil, fmt.Errorf(
				"failed to encode TSS round one message: [%v]",
				err,
			)
		}

		return &tssRoundThreeMessage{
			senderID:  trtm.id,
			payload:   tssMessageBytes,
			sessionID: trtm.sessionID,
		}, nil
	case <-ctx.Done():
		return nil, fmt.Errorf(
			"TSS round three outgoing message was not generated on time",
		)
	}
}

// tssFinalize finalizes the TSS process by producing a result.
func (fm *finalizingMember) tssFinalize(
	ctx context.Context,
	tssRoundThreeMessages []*tssRoundThreeMessage,
) error {
	// Use messages from round three to update the local party and get the
	// result.
	for _, tssRoundThreeMessage := range deduplicateBySender(tssRoundThreeMessages) {
		senderID := tssRoundThreeMessage.SenderID()

		_, tssErr := fm.tssParty.UpdateFromBytes(
			tssRoundThreeMessage.payload,
			common.ResolveSortedTssPartyID(fm.tssParameters, senderID),
			true,
		)
		if tssErr != nil {
			return fmt.Errorf(
				"cannot update using TSS round three message "+
					"from member [%v]: [%v]",
				senderID,
				tssErr,
			)
		}
	}

	select {
	case tssResult := <-fm.tssResultChan:
		fm.tssResult = tssResult
		return nil
	case <-ctx.Done():
		return fmt.Errorf(
			"TSS result was not generated on time",
		)
	}
}

// signDKGResult signs the provided DKG result and prepares the appropriate
// result signature message.
func (sm *signingMember) signDKGResult(
	dkgResult *Result,
	resultSigner ResultSigner,
) (*resultSignatureMessage, error) {
	signedResult, err := resultSigner.SignResult(dkgResult)
	if err != nil {
		return nil, fmt.Errorf("failed to sign DKG result [%v]", err)
	}

	// Register self signature and result hash.
	sm.selfDKGResultSignature = signedResult.Signature
	sm.preferredDKGResultHash = signedResult.ResultHash

	return &resultSignatureMessage{
		senderID:   sm.memberIndex,
		resultHash: signedResult.ResultHash,
		signature:  signedResult.Signature,
		publicKey:  signedResult.PublicKey,
		sessionID:  sm.sessionID,
	}, nil
}

// verifyDKGResultSignatures verifies signatures received in messages from other
// group members. It collects signatures supporting only the same DKG result
// hash as the one preferred by the current member. Each member is allowed to
// broadcast only one signature over a preferred DKG result hash. The function
// assumes that the input messages list does not contain a message from self and
// that the public key presented in each message is the correct one.
// This key needs to be compared against the one used by network client earlier,
// before this function is called.
func (sm *signingMember) verifyDKGResultSignatures(
	messages []*resultSignatureMessage,
	resultSigner ResultSigner,
) map[group.MemberIndex][]byte {
	receivedValidResultSignatures := make(map[group.MemberIndex][]byte)

	for _, message := range deduplicateBySender(messages) {
		// Sender's preferred DKG result hash doesn't match current member's
		// preferred DKG result hash.
		if message.resultHash != sm.preferredDKGResultHash {
			sm.logger.Infof(
				"[member: %v] signature from sender [%d] supports "+
					"result different than preferred",
				sm.memberIndex,
				message.senderID,
			)
			continue
		}

		// Check if the signature is valid.
		isValid, err := resultSigner.VerifySignature(
			&SignedResult{
				ResultHash: message.resultHash,
				Signature:  message.signature,
				PublicKey:  message.publicKey,
			},
		)
		if err != nil {
			sm.logger.Infof(
				"[member: %v] verification of signature "+
					"from sender [%d] failed: [%v]",
				sm.memberIndex,
				message.senderID,
				err,
			)
			continue
		}
		if !isValid {
			sm.logger.Infof(
				"[member: %v] sender [%d] provided invalid signature",
				sm.memberIndex,
				message.senderID,
			)
			continue
		}

		receivedValidResultSignatures[message.senderID] = message.signature
	}

	// Register member's self signature.
	receivedValidResultSignatures[sm.memberIndex] = sm.selfDKGResultSignature

	return receivedValidResultSignatures
}

// submitDKGResult submits the DKG result along with the supporting signatures
// to the provided result submitter.
func (sm *submittingMember) submitDKGResult(
	result *Result,
	signatures map[group.MemberIndex][]byte,
	startBlockNumber uint64,
	resultSubmitter ResultSubmitter,
) error {
	if err := resultSubmitter.SubmitResult(
		sm.memberIndex,
		result,
		signatures,
		startBlockNumber,
	); err != nil {
		return fmt.Errorf("failed to submit DKG result [%v]", err)
	}

	return nil
}

// deduplicateBySender removes duplicated items for the given sender.
// It always takes the first item that occurs for the given sender
// and ignores the subsequent ones.
func deduplicateBySender[T interface{ SenderID() group.MemberIndex }](
	list []T,
) []T {
	senders := make(map[group.MemberIndex]bool)
	result := make([]T, 0)

	for _, item := range list {
		if _, exists := senders[item.SenderID()]; !exists {
			senders[item.SenderID()] = true
			result = append(result, item)
		}
	}

	return result
}
