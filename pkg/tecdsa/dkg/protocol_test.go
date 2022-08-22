package dkg

import (
	"context"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bnb-chain/tss-lib/crypto/paillier"
	"github.com/bnb-chain/tss-lib/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/tss"
	"github.com/keep-network/keep-core/pkg/crypto/ephemeral"
	"github.com/keep-network/keep-core/pkg/internal/testutils"
	"github.com/keep-network/keep-core/pkg/protocol/group"
)

// TODO: This file contains unit tests that stress each protocol phase
//       separately. We should also develop integration tests just like
//       we did for Random Beacon's DKG protocol. This will require
//       to refactor the `pkg/internal/dkgtest` package in the way that
//       it supports the ECDSA DKG as well.

const (
	groupSize          = 3
	dishonestThreshold = 0
	sessionID          = "session-1"
)

func TestGenerateEphemeralKeyPair(t *testing.T) {
	members, err := initializeEphemeralKeyPairGeneratingMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Generate ephemeral key pairs for each group member.
	messages := make(map[group.MemberIndex]*ephemeralPublicKeyMessage)
	for _, member := range members {
		message, err := member.generateEphemeralKeyPair()
		if err != nil {
			t.Fatal(err)
		}
		messages[member.id] = message
	}

	// Assert that each member has a correct state.
	for _, member := range members {
		// Assert the right key pairs count is stored in the member's state.
		expectedKeyPairsCount := groupSize - 1
		actualKeyPairsCount := len(member.ephemeralKeyPairs)
		testutils.AssertIntsEqual(
			t,
			fmt.Sprintf(
				"number of stored ephemeral key pairs for member [%v]",
				member.id,
			),
			expectedKeyPairsCount,
			actualKeyPairsCount,
		)

		// Assert the member does not hold a key pair with itself.
		_, ok := member.ephemeralKeyPairs[member.id]
		if ok {
			t.Errorf(
				"[member:%v] found ephemeral key pair generated to self",
				member.id,
			)
		}

		// Assert key pairs are non-nil.
		for otherMemberID, keyPair := range member.ephemeralKeyPairs {
			if keyPair.PrivateKey == nil {
				t.Errorf(
					"[member:%v] key pair's private key not set for member [%v]",
					member.id,
					otherMemberID,
				)
			}

			if keyPair.PublicKey == nil {
				t.Errorf(
					"[member:%v] key pair's public key not set for member [%v]",
					member.id,
					otherMemberID,
				)
			}
		}
	}

	// Assert that each message is formed correctly.
	for memberID, message := range messages {
		// We should always be the sender of our own messages.
		testutils.AssertIntsEqual(
			t,
			"message sender",
			int(memberID),
			int(message.senderID),
		)

		// We should not generate an ephemeral key for ourselves.
		_, ok := message.ephemeralPublicKeys[memberID]
		if ok {
			t.Errorf("found ephemeral key generated to self")
		}
	}
}

func TestGenerateSymmetricKeys(t *testing.T) {
	members, messages, err := initializeSymmetricKeyGeneratingMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Generate symmetric keys for each group member.
	for _, member := range members {
		var receivedMessages []*ephemeralPublicKeyMessage
		for _, message := range messages {
			if message.senderID != member.id {
				receivedMessages = append(receivedMessages, message)
			}
		}

		err := member.generateSymmetricKeys(receivedMessages)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Assert that each member has a correct state.
	for _, member := range members {
		// Assert the right keys count is stored in the member's state.
		expectedKeysCount := groupSize - 1
		actualKeysCount := len(member.symmetricKeys)
		testutils.AssertIntsEqual(
			t,
			fmt.Sprintf(
				"number of stored symmetric keys for member [%v]",
				member.id,
			),
			expectedKeysCount,
			actualKeysCount,
		)

		// Assert all symmetric keys stored by this member are correct.
		for otherMemberID, actualKey := range member.symmetricKeys {
			var otherMemberEphemeralPublicKey *ephemeral.PublicKey
			for _, message := range messages {
				if message.senderID == otherMemberID {
					if ephemeralPublicKey, ok := message.ephemeralPublicKeys[member.id]; ok {
						otherMemberEphemeralPublicKey = ephemeralPublicKey
					}
				}
			}

			if otherMemberEphemeralPublicKey == nil {
				t.Errorf(
					"[member:%v] no ephemeral public key from member [%v]",
					member.id,
					otherMemberID,
				)
			}

			expectedKey := ephemeral.SymmetricKey(
				member.ephemeralKeyPairs[otherMemberID].PrivateKey.Ecdh(
					otherMemberEphemeralPublicKey,
				),
			)

			if !reflect.DeepEqual(
				expectedKey,
				actualKey,
			) {
				t.Errorf(
					"[member:%v] wrong symmetric key for member [%v]",
					member.id,
					otherMemberID,
				)
			}
		}
	}
}

func TestGenerateSymmetricKeys_InvalidEphemeralPublicKeyMessage(t *testing.T) {
	members, messages, err := initializeSymmetricKeyGeneratingMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Corrupt the message sent by member 2 by removing the ephemeral
	// public key generated for member 3.
	misbehavingMemberID := group.MemberIndex(2)
	delete(messages[misbehavingMemberID-1].ephemeralPublicKeys, 3)

	// Generate symmetric keys for each group member.
	for _, member := range members {
		var receivedMessages []*ephemeralPublicKeyMessage
		for _, message := range messages {
			if message.senderID != member.id {
				receivedMessages = append(receivedMessages, message)
			}
		}

		err := member.generateSymmetricKeys(receivedMessages)

		var expectedErr error
		// The misbehaved member should not get an error.
		if member.id != misbehavingMemberID {
			expectedErr = fmt.Errorf(
				"member [%v] sent invalid ephemeral "+
					"public key message",
				misbehavingMemberID,
			)
		}

		if !reflect.DeepEqual(expectedErr, err) {
			t.Errorf(
				"unexpected error\nexpected: %v\nactual:   %v\n",
				expectedErr,
				err,
			)
		}
	}
}

func TestTssRoundOne(t *testing.T) {
	members, err := initializeTssRoundOneMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS round one for each group member.
	messages := make(map[group.MemberIndex]*tssRoundOneMessage)
	for _, member := range members {
		ctx, cancelCtx := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)

		message, err := member.tssRoundOne(ctx)
		if err != nil {
			cancelCtx()
			t.Fatal(err)
		}
		messages[member.id] = message

		cancelCtx()
	}

	// Assert that each member has a correct state.
	for _, member := range members {
		if !strings.Contains(member.tssParty.String(), "round: 1") {
			t.Errorf("wrong round number for member [%v]", member.id)
		}
	}

	// Assert that each message is formed correctly.
	for memberID, message := range messages {
		// We should always be the sender of our own messages.
		testutils.AssertIntsEqual(
			t,
			fmt.Sprintf(
				"message sender in message generated by member [%v]",
				memberID,
			),
			int(memberID),
			int(message.senderID),
		)

		// We should always generate a payload.
		if len(message.payload) == 0 {
			t.Errorf(
				"empty payload in message generated by member [%v]",
				memberID,
			)
		}

		// We should always use the proper session ID.
		testutils.AssertStringsEqual(
			t,
			fmt.Sprintf(
				"session ID in message generated by member [%v]",
				memberID,
			),
			sessionID,
			message.sessionID,
		)
	}
}

func TestTssRoundOne_OutgoingMessageTimeout(t *testing.T) {
	members, err := initializeTssRoundOneMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS round one for each group member.
	for _, member := range members {
		// To simulate the outgoing message timeout we do two things:
		// - we pass an already cancelled context
		// - we make sure no message is emitted from the channel by overwriting
		//   the existing channel with a new one that won't receive any
		//   messages from the underlying TSS local party
		ctx, cancelCtx := context.WithCancel(context.Background())
		cancelCtx()
		member.tssOutgoingMessagesChan = make(<-chan tss.Message)

		_, err := member.tssRoundOne(ctx)

		expectedErr := fmt.Errorf(
			"TSS round one outgoing message was not generated on time",
		)
		if !reflect.DeepEqual(expectedErr, err) {
			t.Errorf(
				"unexpected error for member [%v]\n"+
					"expected: %v\n"+
					"actual:   %v\n",
				member.id,
				expectedErr,
				err,
			)
		}
	}
}

func TestTssRoundTwo(t *testing.T) {
	members, tssRoundOneMessages, err := initializeTssRoundTwoMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS round two for each group member.
	tssRoundTwoMessages := make(map[group.MemberIndex]*tssRoundTwoMessage)
	for _, member := range members {
		var receivedTssRoundOneMessages []*tssRoundOneMessage
		for _, tssRoundOneMessage := range tssRoundOneMessages {
			if tssRoundOneMessage.senderID != member.id {
				receivedTssRoundOneMessages = append(
					receivedTssRoundOneMessages,
					tssRoundOneMessage,
				)
			}
		}

		ctx, cancelCtx := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)

		tssRoundTwoMessage, err := member.tssRoundTwo(
			ctx,
			receivedTssRoundOneMessages,
		)
		if err != nil {
			cancelCtx()
			t.Fatal(err)
		}
		tssRoundTwoMessages[member.id] = tssRoundTwoMessage

		cancelCtx()
	}

	// Assert that each member has a correct state.
	for _, member := range members {
		if !strings.Contains(member.tssParty.String(), "round: 2") {
			t.Errorf("wrong round number for member [%v]", member.id)
		}
	}

	// Assert that each message is formed correctly.
	for memberID, message := range tssRoundTwoMessages {
		// We should always be the sender of our own messages.
		testutils.AssertIntsEqual(
			t,
			fmt.Sprintf(
				"message sender in message generated by member [%v]",
				memberID,
			),
			int(memberID),
			int(message.senderID),
		)

		// We should always generate a broadcast payload.
		if len(message.broadcastPayload) == 0 {
			t.Errorf(
				"empty broadcast payload in message generated by member [%v]",
				memberID,
			)
		}

		// We should always generate groupSize-1 of peers payloads.
		testutils.AssertIntsEqual(
			t,
			fmt.Sprintf(
				"count of peers payloads in message "+
					"generated by member [%v]",
				memberID,
			),
			groupSize-1,
			len(message.peersPayload),
		)

		// Each P2P payload should be encrypted using the proper symmetric key.
		for receiverID, encryptedPayload := range message.peersPayload {
			symmetricKey := members[memberID-1].symmetricKeys[receiverID]
			if _, err := symmetricKey.Decrypt(encryptedPayload); err != nil {
				t.Errorf(
					"payload for member [%v] in message generated "+
						"by member [%v] is encrypted using "+
						"the wrong symmetric key: [%v]",
					receiverID,
					memberID,
					err,
				)
			}
		}

		// We should always use the proper session ID.
		testutils.AssertStringsEqual(
			t,
			fmt.Sprintf(
				"session ID in message generated by member [%v]",
				memberID,
			),
			sessionID,
			message.sessionID,
		)
	}
}

func TestTssRoundTwo_IncomingMessageCorrupted_WrongPayload(t *testing.T) {
	members, messages, err := initializeTssRoundTwoMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	corruptedPayload, err := hex.DecodeString("ffeeaabb")
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS round two for each group member.
	for _, member := range members {
		var receivedMessages []*tssRoundOneMessage
		for _, message := range messages {
			if message.senderID != member.id {
				// Corrupt the message's payload.
				message.payload = corruptedPayload
				receivedMessages = append(receivedMessages, message)
			}
		}

		ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)

		_, err := member.tssRoundTwo(ctx, receivedMessages)

		if !strings.Contains(
			err.Error(),
			"cannot update using TSS round one message",
		) {
			t.Errorf("wrong error for member [%v]: [%v]", member.id, err)
		}

		cancelCtx()
	}
}

func TestTssRoundTwo_IncomingMessageMissing(t *testing.T) {
	members, messages, err := initializeTssRoundTwoMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS round two for each group member.
	for _, member := range members {
		var receivedMessages []*tssRoundOneMessage
		for _, message := range messages {
			if message.senderID != member.id {
				receivedMessages = append(receivedMessages, message)
			}
		}

		ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Second)
		// Pass only one incoming message from TSS round one for processing.
		_, err := member.tssRoundTwo(ctx, receivedMessages[:1])

		expectedErr := fmt.Errorf(
			"TSS round two outgoing messages were not generated on time",
		)
		if !reflect.DeepEqual(expectedErr, err) {
			t.Errorf(
				"unexpected error for member [%v]\n"+
					"expected: %v\n"+
					"actual:   %v\n",
				member.id,
				expectedErr,
				err,
			)
		}

		cancelCtx()
	}
}

func TestTssRoundTwo_OutgoingMessageTimeout(t *testing.T) {
	members, messages, err := initializeTssRoundTwoMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS round two for each group member.
	for _, member := range members {
		var receivedMessages []*tssRoundOneMessage
		for _, message := range messages {
			if message.senderID != member.id {
				receivedMessages = append(receivedMessages, message)
			}
		}

		// To simulate the outgoing message timeout we do two things:
		// - we pass an already cancelled context
		// - we make sure no message is emitted from the channel by overwriting
		//   the existing channel with a new one that won't receive any
		//   messages from the underlying TSS local party
		ctx, cancelCtx := context.WithCancel(context.Background())
		cancelCtx()
		member.tssOutgoingMessagesChan = make(<-chan tss.Message)

		_, err := member.tssRoundTwo(ctx, receivedMessages)

		expectedErr := fmt.Errorf(
			"TSS round two outgoing messages were not generated on time",
		)
		if !reflect.DeepEqual(expectedErr, err) {
			t.Errorf(
				"unexpected error for member [%v]\n"+
					"expected: %v\n"+
					"actual:   %v\n",
				member.id,
				expectedErr,
				err,
			)
		}
	}
}

func TestTssRoundTwo_SymmetricKeyMissing(t *testing.T) {
	members, messages, err := initializeTssRoundTwoMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS round two for each group member.
	for _, member := range members {
		var receivedMessages []*tssRoundOneMessage
		for _, message := range messages {
			if message.senderID != member.id {
				receivedMessages = append(receivedMessages, message)
			}
		}

		// Cleanup symmetric key cache.
		member.symmetricKeys = make(map[group.MemberIndex]ephemeral.SymmetricKey)

		ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)

		_, err := member.tssRoundTwo(ctx, receivedMessages)

		if !strings.Contains(
			err.Error(),
			"cannot get symmetric key with member",
		) {
			t.Errorf("wrong error for member [%v]: [%v]", member.id, err)
		}

		cancelCtx()
	}
}

func TestTssRoundThree(t *testing.T) {
	members, tssRoundTwoMessages, err := initializeTssRoundThreeMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS round three for each group member.
	tssRoundThreeMessages := make(map[group.MemberIndex]*tssRoundThreeMessage)
	for _, member := range members {
		var receivedTssRoundTwoMessages []*tssRoundTwoMessage
		for _, tssRoundTwoMessage := range tssRoundTwoMessages {
			if tssRoundTwoMessage.senderID != member.id {
				receivedTssRoundTwoMessages = append(
					receivedTssRoundTwoMessages,
					tssRoundTwoMessage,
				)
			}
		}

		ctx, cancelCtx := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)

		tssRoundThreeMessage, err := member.tssRoundThree(
			ctx,
			receivedTssRoundTwoMessages,
		)
		if err != nil {
			cancelCtx()
			t.Fatal(err)
		}
		tssRoundThreeMessages[member.id] = tssRoundThreeMessage

		cancelCtx()
	}

	// Assert that each member has a correct state.
	for _, member := range members {
		if !strings.Contains(member.tssParty.String(), "round: 3") {
			t.Errorf("wrong round number for member [%v]", member.id)
		}
	}

	// Assert that each message is formed correctly.
	for memberID, message := range tssRoundThreeMessages {
		// We should always be the sender of our own messages.
		testutils.AssertIntsEqual(
			t,
			fmt.Sprintf(
				"message sender in message generated by member [%v]",
				memberID,
			),
			int(memberID),
			int(message.senderID),
		)

		// We should always generate a payload.
		if len(message.payload) == 0 {
			t.Errorf(
				"empty payload in message generated by member [%v]",
				memberID,
			)
		}

		// We should always use the proper session ID.
		testutils.AssertStringsEqual(
			t,
			fmt.Sprintf(
				"session ID in message generated by member [%v]",
				memberID,
			),
			sessionID,
			message.sessionID,
		)
	}
}

func TestTssRoundTwo_IncomingMessageCorrupted_WrongBroadcastPayload(t *testing.T) {
	members, messages, err := initializeTssRoundThreeMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	corruptedPayload, err := hex.DecodeString("ffeeaabb")
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS round three for each group member.
	for _, member := range members {
		var receivedMessages []*tssRoundTwoMessage
		for _, message := range messages {
			if message.senderID != member.id {
				// Corrupt the message's broadcast payload.
				message.broadcastPayload = corruptedPayload
				receivedMessages = append(receivedMessages, message)
			}
		}

		ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)

		_, err := member.tssRoundThree(ctx, receivedMessages)

		if !strings.Contains(
			err.Error(),
			"cannot update using the broadcast part of the TSS round two message",
		) {
			t.Errorf("wrong error for member [%v]: [%v]", member.id, err)
		}

		cancelCtx()
	}
}

func TestTssRoundTwo_IncomingMessageCorrupted_UndecryptablePeerPayload(t *testing.T) {
	members, messages, err := initializeTssRoundThreeMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	corruptedPayload, err := hex.DecodeString("ffeeaabb")
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS round three for each group member.
	for _, member := range members {
		var receivedMessages []*tssRoundTwoMessage
		for _, message := range messages {
			if message.senderID != member.id {
				// Make the P2P undecryptable by setting an arbitrary value
				// as ciphertext.
				corruptedPeersPayload := make(map[group.MemberIndex][]byte)
				for receiverID := range message.peersPayload {
					corruptedPeersPayload[receiverID] = corruptedPayload
				}
				message.peersPayload = corruptedPeersPayload
				receivedMessages = append(receivedMessages, message)
			}
		}

		ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)

		_, err := member.tssRoundThree(ctx, receivedMessages)

		if !strings.Contains(
			err.Error(),
			"cannot decrypt P2P part of the TSS round two message",
		) {
			t.Errorf("wrong error for member [%v]: [%v]", member.id, err)
		}

		cancelCtx()
	}
}

func TestTssRoundTwo_IncomingMessageCorrupted_WrongPeerPayload(t *testing.T) {
	members, messages, err := initializeTssRoundThreeMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	corruptedPayload, err := hex.DecodeString("ffeeaabb")
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS round three for each group member.
	for _, member := range members {
		var receivedMessages []*tssRoundTwoMessage
		for _, message := range messages {
			if message.senderID != member.id {
				// Corrupt the message's peers payload by encrypting an
				// arbitrary value.
				corruptedPeersPayload := make(map[group.MemberIndex][]byte)
				for receiverID := range message.peersPayload {
					symmetricKey := members[message.senderID-1].symmetricKeys[receiverID]
					encryptedCorruptedPayload, err := symmetricKey.Encrypt(corruptedPayload)
					if err != nil {
						t.Fatal(err)
					}
					corruptedPeersPayload[receiverID] = encryptedCorruptedPayload
				}
				message.peersPayload = corruptedPeersPayload
				receivedMessages = append(receivedMessages, message)
			}
		}

		ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)

		_, err := member.tssRoundThree(ctx, receivedMessages)

		if !strings.Contains(
			err.Error(),
			"cannot update using the P2P part of the TSS round two message",
		) {
			t.Errorf("wrong error for member [%v]: [%v]", member.id, err)
		}

		cancelCtx()
	}
}

func TestTssRoundThree_IncomingMessageMissing(t *testing.T) {
	members, messages, err := initializeTssRoundThreeMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS round three for each group member.
	for _, member := range members {
		var receivedMessages []*tssRoundTwoMessage
		for _, message := range messages {
			if message.senderID != member.id {
				receivedMessages = append(receivedMessages, message)
			}
		}

		ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Second)
		// Pass only one incoming message from TSS round two for processing.
		_, err := member.tssRoundThree(ctx, receivedMessages[:1])

		expectedErr := fmt.Errorf(
			"TSS round three outgoing message was not generated on time",
		)
		if !reflect.DeepEqual(expectedErr, err) {
			t.Errorf(
				"unexpected error for member [%v]\n"+
					"expected: %v\n"+
					"actual:   %v\n",
				member.id,
				expectedErr,
				err,
			)
		}

		cancelCtx()
	}
}

func TestTssRoundThree_OutgoingMessageTimeout(t *testing.T) {
	members, messages, err := initializeTssRoundThreeMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS round three for each group member.
	for _, member := range members {
		var receivedMessages []*tssRoundTwoMessage
		for _, message := range messages {
			if message.senderID != member.id {
				receivedMessages = append(receivedMessages, message)
			}
		}

		// To simulate the outgoing message timeout we do two things:
		// - we pass an already cancelled context
		// - we make sure no message is emitted from the channel by overwriting
		//   the existing channel with a new one that won't receive any
		//   messages from the underlying TSS local party
		ctx, cancelCtx := context.WithCancel(context.Background())
		cancelCtx()
		member.tssOutgoingMessagesChan = make(<-chan tss.Message)

		_, err := member.tssRoundThree(ctx, receivedMessages)

		expectedErr := fmt.Errorf(
			"TSS round three outgoing message was not generated on time",
		)
		if !reflect.DeepEqual(expectedErr, err) {
			t.Errorf(
				"unexpected error for member [%v]\n"+
					"expected: %v\n"+
					"actual:   %v\n",
				member.id,
				expectedErr,
				err,
			)
		}
	}
}

func TestTssRoundThree_SymmetricKeyMissing(t *testing.T) {
	members, messages, err := initializeTssRoundThreeMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS round two for each group member.
	for _, member := range members {
		var receivedMessages []*tssRoundTwoMessage
		for _, message := range messages {
			if message.senderID != member.id {
				receivedMessages = append(receivedMessages, message)
			}
		}

		// Cleanup symmetric key cache.
		member.symmetricKeys = make(map[group.MemberIndex]ephemeral.SymmetricKey)

		ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)

		_, err := member.tssRoundThree(ctx, receivedMessages)

		if !strings.Contains(
			err.Error(),
			"cannot get symmetric key with member",
		) {
			t.Errorf("wrong error for member [%v]: [%v]", member.id, err)
		}

		cancelCtx()
	}
}

func TestTssFinalize(t *testing.T) {
	members, tssRoundThreeMessages, err := initializeFinalizingMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS finalization for each group member.
	for _, member := range members {
		var receivedTssRoundThreeMessages []*tssRoundThreeMessage
		for _, tssRoundThreeMessage := range tssRoundThreeMessages {
			if tssRoundThreeMessage.senderID != member.id {
				receivedTssRoundThreeMessages = append(
					receivedTssRoundThreeMessages,
					tssRoundThreeMessage,
				)
			}
		}

		ctx, cancelCtx := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)

		err := member.tssFinalize(
			ctx,
			receivedTssRoundThreeMessages,
		)
		if err != nil {
			cancelCtx()
			t.Fatal(err)
		}

		cancelCtx()
	}

	groupPublicKeys := make(map[string]bool)

	// Assert that each member has a correct state.
	for _, member := range members {
		if member.Result().PrivateKeyShare == nil {
			t.Fatalf(
				"member [%v] has not produced a private key share",
				member.id,
			)
		}

		groupPublicKey := member.Result().PrivateKeyShare.PublicKey()

		groupPublicKeyBytes := elliptic.Marshal(
			groupPublicKey.Curve,
			groupPublicKey.X,
			groupPublicKey.Y,
		)

		groupPublicKeys[hex.EncodeToString(groupPublicKeyBytes)] = true
	}

	testutils.AssertIntsEqual(
		t,
		"count of distinct group public keys produced by the group",
		1,
		len(groupPublicKeys),
	)
}

func TestTssFinalize_IncomingMessageCorrupted_WrongPayload(t *testing.T) {
	members, messages, err := initializeFinalizingMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	corruptedPayload, err := hex.DecodeString("ffeeaabb")
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS round two for each group member.
	for _, member := range members {
		var receivedMessages []*tssRoundThreeMessage
		for _, message := range messages {
			if message.senderID != member.id {
				// Corrupt the message's payload.
				message.payload = corruptedPayload
				receivedMessages = append(receivedMessages, message)
			}
		}

		ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)

		err := member.tssFinalize(ctx, receivedMessages)

		if !strings.Contains(
			err.Error(),
			"cannot update using TSS round three message",
		) {
			t.Errorf("wrong error for member [%v]: [%v]", member.id, err)
		}

		cancelCtx()
	}
}

func TestTssFinalize_IncomingMessageMissing(t *testing.T) {
	members, messages, err := initializeFinalizingMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS round two for each group member.
	for _, member := range members {
		var receivedMessages []*tssRoundThreeMessage
		for _, message := range messages {
			if message.senderID != member.id {
				receivedMessages = append(receivedMessages, message)
			}
		}

		ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Second)
		// Pass only one incoming message from TSS round three for processing.
		err := member.tssFinalize(ctx, receivedMessages[:1])

		expectedErr := fmt.Errorf(
			"TSS result was not generated on time",
		)
		if !reflect.DeepEqual(expectedErr, err) {
			t.Errorf(
				"unexpected error for member [%v]\n"+
					"expected: %v\n"+
					"actual:   %v\n",
				member.id,
				expectedErr,
				err,
			)
		}

		cancelCtx()
	}
}

func TestTssFinalize_ResultTimeout(t *testing.T) {
	members, messages, err := initializeFinalizingMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS finalization for each group member.
	for _, member := range members {
		var receivedMessages []*tssRoundThreeMessage
		for _, message := range messages {
			if message.senderID != member.id {
				receivedMessages = append(receivedMessages, message)
			}
		}

		// To simulate the outgoing message timeout we do two things:
		// - we pass an already cancelled context
		// - we make sure no result is emitted from the channel by overwriting
		//   the existing channel with a new one that won't receive the
		//   result from the underlying TSS local party
		ctx, cancelCtx := context.WithCancel(context.Background())
		cancelCtx()
		member.tssResultChan = make(<-chan keygen.LocalPartySaveData)

		err := member.tssFinalize(ctx, receivedMessages)

		expectedErr := fmt.Errorf(
			"TSS result was not generated on time",
		)
		if !reflect.DeepEqual(expectedErr, err) {
			t.Errorf(
				"unexpected error for member [%v]\n"+
					"expected: %v\n"+
					"actual:   %v\n",
				member.id,
				expectedErr,
				err,
			)
		}
	}
}

func initializeEphemeralKeyPairGeneratingMembersGroup(
	dishonestThreshold int,
	groupSize int,
) ([]*ephemeralKeyPairGeneratingMember, error) {
	dkgGroup := group.NewGroup(dishonestThreshold, groupSize)

	tssPreParams, err := generateMembersTssPreParams(groupSize)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot generate members TSS pre-parameters: [%v]",
			err,
		)
	}

	var members []*ephemeralKeyPairGeneratingMember
	for i := 1; i <= groupSize; i++ {
		id := group.MemberIndex(i)

		members = append(members, &ephemeralKeyPairGeneratingMember{
			member: &member{
				logger:       &testutils.MockLogger{},
				id:           id,
				group:        dkgGroup,
				sessionID:    sessionID,
				tssPreParams: tssPreParams[id],
			},
			ephemeralKeyPairs: make(map[group.MemberIndex]*ephemeral.KeyPair),
		})
	}

	return members, nil
}
func initializeSymmetricKeyGeneratingMembersGroup(
	dishonestThreshold int,
	groupSize int,
) (
	[]*symmetricKeyGeneratingMember,
	[]*ephemeralPublicKeyMessage,
	error,
) {
	var symmetricKeyGeneratingMembers []*symmetricKeyGeneratingMember
	var ephemeralPublicKeyMessages []*ephemeralPublicKeyMessage

	ephemeralKeyPairGeneratingMembers, err :=
		initializeEphemeralKeyPairGeneratingMembersGroup(
			dishonestThreshold,
			groupSize,
		)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"cannot generate ephemeral key pair generating "+
				"members group: [%v]",
			err,
		)
	}

	for _, member := range ephemeralKeyPairGeneratingMembers {
		message, err := member.generateEphemeralKeyPair()
		if err != nil {
			return nil, nil, fmt.Errorf(
				"cannot generate ephemeral key pair for member [%v]: [%v]",
				member.id,
				err,
			)
		}

		symmetricKeyGeneratingMembers = append(
			symmetricKeyGeneratingMembers,
			member.initializeSymmetricKeyGeneration(),
		)
		ephemeralPublicKeyMessages = append(ephemeralPublicKeyMessages, message)
	}

	return symmetricKeyGeneratingMembers, ephemeralPublicKeyMessages, nil
}

func initializeTssRoundOneMembersGroup(
	dishonestThreshold int,
	groupSize int,
) ([]*tssRoundOneMember, error) {
	var tssRoundOneMembers []*tssRoundOneMember

	symmetricKeyGeneratingMembers, ephemeralPublicKeyMessages, err :=
		initializeSymmetricKeyGeneratingMembersGroup(
			dishonestThreshold,
			groupSize,
		)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot generate symmetric key generating members group: [%v]",
			err,
		)
	}

	for _, member := range symmetricKeyGeneratingMembers {
		var receivedMessages []*ephemeralPublicKeyMessage
		for _, message := range ephemeralPublicKeyMessages {
			if message.senderID != member.id {
				receivedMessages = append(receivedMessages, message)
			}
		}

		err := member.generateSymmetricKeys(receivedMessages)
		if err != nil {
			return nil, fmt.Errorf(
				"cannot generate symmetric keys for member [%v]: [%v]",
				member.id,
				err,
			)
		}

		tssRoundOneMembers = append(
			tssRoundOneMembers,
			member.initializeTssRoundOne(),
		)
	}

	return tssRoundOneMembers, nil
}

func initializeTssRoundTwoMembersGroup(
	dishonestThreshold int,
	groupSize int,
) (
	[]*tssRoundTwoMember,
	[]*tssRoundOneMessage,
	error,
) {
	var tssRoundTwoMembers []*tssRoundTwoMember
	var tssRoundOneMessages []*tssRoundOneMessage

	tssRoundOneMembers, err :=
		initializeTssRoundOneMembersGroup(
			dishonestThreshold,
			groupSize,
		)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"cannot generate TSS round one members group: [%v]",
			err,
		)
	}

	for _, member := range tssRoundOneMembers {
		ctx, cancelCtx := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)

		message, err := member.tssRoundOne(ctx)
		if err != nil {
			cancelCtx()
			return nil, nil, fmt.Errorf(
				"cannot do TSS round one for member [%v]: [%v]",
				member.id,
				err,
			)
		}

		tssRoundTwoMembers = append(
			tssRoundTwoMembers,
			member.initializeTssRoundTwo(),
		)
		tssRoundOneMessages = append(tssRoundOneMessages, message)

		cancelCtx()
	}

	return tssRoundTwoMembers, tssRoundOneMessages, nil
}

func initializeTssRoundThreeMembersGroup(
	dishonestThreshold int,
	groupSize int,
) (
	[]*tssRoundThreeMember,
	[]*tssRoundTwoMessage,
	error,
) {
	var tssRoundThreeMembers []*tssRoundThreeMember
	var tssRoundTwoMessages []*tssRoundTwoMessage

	tssRoundTwoMembers, tssRoundOneMessages, err :=
		initializeTssRoundTwoMembersGroup(
			dishonestThreshold,
			groupSize,
		)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"cannot generate TSS round two members group: [%v]",
			err,
		)
	}

	for _, member := range tssRoundTwoMembers {
		var receivedTssRoundOneMessages []*tssRoundOneMessage
		for _, tssRoundOneMessage := range tssRoundOneMessages {
			if tssRoundOneMessage.senderID != member.id {
				receivedTssRoundOneMessages = append(
					receivedTssRoundOneMessages,
					tssRoundOneMessage,
				)
			}
		}

		ctx, cancelCtx := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)

		tssRoundTwoMessage, err := member.tssRoundTwo(
			ctx,
			receivedTssRoundOneMessages,
		)
		if err != nil {
			cancelCtx()
			return nil, nil, fmt.Errorf(
				"cannot do TSS round two for member [%v]: [%v]",
				member.id,
				err,
			)
		}

		tssRoundThreeMembers = append(
			tssRoundThreeMembers,
			member.initializeTssRoundThree(),
		)
		tssRoundTwoMessages = append(tssRoundTwoMessages, tssRoundTwoMessage)

		cancelCtx()
	}

	return tssRoundThreeMembers, tssRoundTwoMessages, nil
}

func initializeFinalizingMembersGroup(
	dishonestThreshold int,
	groupSize int,
) (
	[]*finalizingMember,
	[]*tssRoundThreeMessage,
	error,
) {
	var finalizingMembers []*finalizingMember
	var tssRoundThreeMessages []*tssRoundThreeMessage

	tssRoundThreeMembers, tssRoundTwoMessages, err :=
		initializeTssRoundThreeMembersGroup(
			dishonestThreshold,
			groupSize,
		)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"cannot generate TSS round three members group: [%v]",
			err,
		)
	}

	for _, member := range tssRoundThreeMembers {
		var receivedTssRoundTwoMessages []*tssRoundTwoMessage
		for _, tssRoundTwoMessage := range tssRoundTwoMessages {
			if tssRoundTwoMessage.senderID != member.id {
				receivedTssRoundTwoMessages = append(
					receivedTssRoundTwoMessages,
					tssRoundTwoMessage,
				)
			}
		}

		ctx, cancelCtx := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)

		tssRoundThreeMessage, err := member.tssRoundThree(
			ctx,
			receivedTssRoundTwoMessages,
		)
		if err != nil {
			cancelCtx()
			return nil, nil, fmt.Errorf(
				"cannot do TSS round three for member [%v]: [%v]",
				member.id,
				err,
			)
		}

		finalizingMembers = append(
			finalizingMembers,
			member.initializeFinalization(),
		)
		tssRoundThreeMessages = append(tssRoundThreeMessages, tssRoundThreeMessage)

		cancelCtx()
	}

	return finalizingMembers, tssRoundThreeMessages, nil
}

func generateMembersTssPreParams(
	groupSize int,
) (map[group.MemberIndex]*keygen.LocalPreParams, error) {
	predefinedMembersTssPreParams := map[group.MemberIndex]*keygen.LocalPreParams{
		1: newTssPreParams(
			"25922769748919102678415192880711636156565612427571550685296776086119205445525743826557545692077634738129321690187868055737306626420419536394422682260657759329710259802294458956279773225258250955469954464209933873407784778802101265717840506851919529598154066919091078766953942869622551929743069097967501533345363150709912011028449270819442207860620552088412428865900112120786495620291333470644949767300948329241775121748888220588626655915013364614554467190860190736954650967874940702908395331234632114014125372505065096924932509595285205788545338407476139436404463823043865599023326570565049384032977060875483209339089",
			"12961384874459551339207596440355818078282806213785775342648388043059602722762871913278772846038817369064660845093934027868653313210209768197211341130328879664855129901147229478139886612629125477734977232104966936703892389401050632858920253425959764799077033459545539383476971434811275964871534548983750766672520115861254316608127511715120909186915818876509880056231208052258262510380080295105153942894215245396124765560528098088543032820032983199681389377630502693810272249886420412628917630701692773559849432356251989417662290420554742302877434371102841200978891107281847266690850557956285688970415890246967698012978",
			"25922769748919102678415192880711636156565612427571550685296776086119205445525743826557545692077634738129321690187868055737306626420419536394422682260657759329710259802294458956279773225258250955469954464209933873407784778802101265717840506851919529598154066919091078766953942869622551929743069097967501533345040231722508633216255023430241818373831637753019760112462416104516525020760160590210307885788430490792249531121056196177086065640065966399362778755261005387620544499772840825257835261403385547119698864712503978835324580841109484605754868742205682401957782214563694533381701115912571377940831780493935396025956",
			"20539613942852364097890357541124859329931817468396278432713468646303963073659662742703665137736867247354367523800071318544570641421320510992705137876681425752810096966415479528824625129989063402576946505816887222102561441464103605308386975248012283762854115939987945603503283072741824666735245204091384515192454349252950007899626081034649919068642018312817079235168086885705851677572363277983076857313399016624874649811334825694862350059490166759704819411086564625186038339099281295128259092469609539775245598320922394808913338827772001777479207381548603315272620456484970681705115865233047669675602308688791376160589",
			"16370062914568124684409954423220013634799944354368183091925443712820668316759795091290952642141219645055533606292548565759917746455430426634828957426644826424037530474618159463204943752577732484149675671820306363344833458247384057865310742915406677379586789735200748327711872632191061145184949312294612467345847214916930759229195852858849386686352293049987465485866498220082468131280135383612600619493426252446949294373638968518891137429993551161437309269629260378927918725566711632082553316166822070110359114229533322390061282040482480263995079579444943917107997110057038662405191417861817663789094790962966996587522",
			"9653640790649475435050720061635061544335995170813227062007808546473167610366804040613054457009646767723479128021709179513573358845884462519136809844401815066012655857973373223748942767836422506840658738556503260986697250346171921063441485400421533124068250604530993514803166454504801884882297625678932746326066096923436475087338628767636689481829832307623108408425959669915171224014581673426602770656342925462023157550194457295116217893440581116140543598050947318929500123378985275492765280831578803707538206440354119287576298034238031692982504012470196898579719660373199491817717767711160029710911173725338539566802",
			"7073137964546302519426197108795918903355600790936955717923736840490732786295482817546181286885485705259790301469527483584427669945842799314651770055406853852732275734013259522600331726874819141516989629984376313964484821900473954306398017682999954174229504658528063236651987893368454589560813095145972845549239634160410038395430555137183455161117726890513476626495652520344277700761372318531991732442923029695918379854101514426650142318224874474725980008331402893988707190510778836547424547034272241175095182865500366323417396500881575168432154152735186187611899531856007795602612877814298712246199637908092122376599",
			"3781329124778805698135968627168562375518994304682072891270490638565903398815921016162668916726922361980836207877243235819696431386786182993335163149600007435177418332221723597286704926041137399234960148965357193156048240916422570358759272483135455734754411862846510585609034734348827060942031583563018534961341701886811090344356131522527105061847489994982866503647681842337628828003678510975110957502599684605684107010421879806261007404227541546063432272552677635838137548724398528410425454871871886019590411197377178358695658433582626781842015808404610705006039671313159344095133029530767173526789878122576408306961",
			"74682834361593481810023372364901068308987897092071652855871125748072435148922776019438622439256305770842874061361575542291646920690415340075876934694226424600012724116678454328570534454824214148215599456938258720769517493756120783418012627126126134213332959856945407004391730249681744107119504561374114258903",
			"68756140947346998421359710991066203811047516135270426947846872112839728963199219078428991392825847385895934224515769850480824784212201628242055180420737700856064230455409099120643105577895376414218231855017888457591638558485545668540344498197465720699981681711132445684644446855255773133247109821728609375713",
		),
		2: newTssPreParams(
			"23930233287283899271771864413305422456138957780711273892670074191715648409585503033095084345383391541524625291548041741990557564183855401706042293717552023237439032182637019639795919249455653535670614575331737610284863144094845900714497635996654401300216924764570210541950557336240993007183309433063094227377624710274228010652758134777897718742178998545079447283838099902510469006366469099975469096355736757507201973304413688395278990349533350163833514531655073848517781662614171483003731680841330633223244205178982328422170273570503713081265847261211618499950287557687314846590616484106774575999250148317390509484773",
			"11965116643641949635885932206652711228069478890355636946335037095857824204792751516547542172691695770762312645774020870995278782091927700853021146858776011618719516091318509819897959624727826767835307287665868805142431572047422950357248817998327200650108462382285105270975278668120496503591654716531547113688656751738100540554882657085156055277574738768917123952100316530725550908644747452231188839176032305030675216428579800880622045367926074519576691554576481478819003213525881378532266888732131750479170727259561318783487243810779611492259253171062407221253763092683460860198262672146170687159308412065832323010774",
			"23930233287283899271771864413305422456138957780711273892670074191715648409585503033095084345383391541524625291548041741990557564183855401706042293717552023237439032182637019639795919249455653535670614575331737610284863144094845900714497635996654401300216924764570210541950557336240993007183309433063094227377313503476201081109765314170312110555149477537834247904200633061451101817289494904462377678352064610061350432857159601761244090735852149039153383109152962957638006427051762757064533777464263500958341454519122637566974487621559222984518506342124814442507526185366921720396525344292341374318616824131664646021548",
			"19461028678249357721701139019984545699598216253588699892259672060166427273458875608319855785678884811755179389274380053495578644060470229307987007292965327985966772681212738091909180148035785695413643708212165777295662698493311553457174395686873169155288384255670661532430410131045712913078128214239252258473814281283319061613409102410606683119900924722782015902970301519339718368508022893331969649513655635811522767629123667744907556474126774472529158147258343482417188228144974952598132795041139358631852141986745214674779692377899411672630850213748161088638857089501019216868292821676374914063004957409393293909513",
			"5792666313208572350705907594949414590804636531753541567087068453778543363077542142305947911768781865374122070750607847515026168979710170113788718287465368491981654823945993941924700096393523941715256095048825025361038086570514643251828085846380954791657627403414038681940866434304045104130214177360598208974353162505514170835103706979081795485247158115653501838694614859268884296606546104394637012764653804556264770973741677326601115655746125293204398034469241183574629519235451142797709676366285591723984602961129858687877266469708766607187000988412118886394000517917001485137799190522482532376333362935442751367745",
			"3578111860663702772408903345930659472256129868015762875031051677614699117364424442270785915866444756532836287879751640816575659073006676007210405202315496945346450727441553016187592354415793891110773645408147683476571812485850037389853330648238106038729525075512542830213816094853869971661719959033499816133612736102442725651388405183329714325258711655979055386253406319598230253658818466953001815116530962213661362799968355793928849708876651937113231862631691372187008559216884922263381652908899796744393080985272287681952842592674663712445064149288160957033524932550150413325097150870149209345404214256294282382085",
			"11421071720691985233805931440613725920647192840657306004580627664269869560827359450311849288274035445082830308979661706753108048712766307196692431931830063360703978457192789078600088154454453923447528766525917272869231453811378073992003872441820913146901968708891167890273179426508931612939315210420141536662858066164555270207634588140969207306567894980319084359328176222470944215148990545719044600126162705080135986981147555994772657029718950668109976569450187736404100221957582390492136834255108985799466784705512624569497708820941678094557313288115000789823971847930771259280468016904034040650165223964465502742163",
			"4227530880403545440038476922105110261199537465641999998158148101603915605231937688109728431992553740717999358378275263490543703125096499417543757592701897468657628484096430689758261242530422788974112689990440998814106644208884422045041519433217236005763161899441271328933810953174242937025629868477234469201962721397475711644378280986530924591451805074246910811173744333178474233373419231266135117437995644802581087590961012020226698544189434580409698374958610263938482416302069628007660903039729910120878609160294838371096166505152238228269549996232209744466071160336256098002053827877135524242668985610586518309037",
			"69148727022965490791371353344188179096380205628223578363347845412422456543356557096766868185510445198293123561559310892475561895225747211013646582990901153064187509486206820830686069751777027208039340667218028595066917519852526144012481411056284833957995414832693297161887219241384165301228065443379819309633",
			"70359316491054176535495692674347777118526215106212851796111499381496549337314923687504495752117695325055665155415615099586640262752041948937762394634356663125386298530536960357064450956605539549770693624698845882927874489981293210912747793431816752540230899424105324436139249872533904838907087453451890367169",
		),
		3: newTssPreParams(
			"23804125140052077689856128298352557083678652474445385365228110453726681237860799979845611556170894187976654278582576364089033396218674226546868809651353049956675922595541689542576794678062495339422204984765419389268325283682512000995221750412104207394441438666051694475950049774094896290106430636216894744335784327798634247450687264677393229214665686649911456587168142148024558282134024448427550922487022680890892554782651383972136386958126051377715096556862662265886688077689941967157694195467190297477735450118736949849327358586935699405848605265912107169200547464609552395233560924746135866463084686118233592906569",
			"11902062570026038844928064149176278541839326237222692682614055226863340618930399989922805778085447093988327139291288182044516698109337113273434404825676524978337961297770844771288397339031247669711102492382709694634162641841256000497610875206052103697220719333025847237975024887047448145053215318108447372167737702441251565075594802542407391503289364676930046848592889764444473208311524176750345361303715776188892120559308145482690640380386008098215824607516345906020591797099324696931930539129840295608183302807198936106271490859318518243627618367482381944500401418308985572354562059638591289588383179774841227539522",
			"23804125140052077689856128298352557083678652474445385365228110453726681237860799979845611556170894187976654278582576364089033396218674226546868809651353049956675922595541689542576794678062495339422204984765419389268325283682512000995221750412104207394441438666051694475950049774094896290106430636216894744335475404882503130151189605084814783006578729353860093697185779528888946416623048353500690722607431552377784241118616290965381280760772016196431649215032691812041183594198649393863861078259680591216366605614397872212542981718637036487255236734964763889000802836617971144709124119277182579176766359549682455079044",
			"23815206664659393600414832732918591362081086959256855451108811883313935088830793690110550688160373127903180149093000695761674277348327575728255258492470452704258920461298225437641154249481888087192237143947805411796310656512191138629555279666557122333244803756577286887501632314162770617970064401783626962319950524158923845138939649762251756759762119774585338772559055859463599094869423262313306255644927649977403492926253217608523813644206820059309357940964633363130901166057002430269910921882664166860038861390305316020579398429144038386189480114288127704265879389663380565983482028227028306457603727009698486364281",
			"14510201356793997359892744405553071944121896518459738320470368478827891282273167297002903311912769777071155241288755372185351193850631471716718530488323104261827697027757019895835360274243188719367049075501436153398857359555924247334095665350350441220453460409146684994664351725204034521761578947269005519140498383255606322844603919000682223940913285551513356600061526959181206093504072618536296265435830192715190515397064435874311583709516596584871136822289753837472532490128304199643440789002058080030111113389709287097531544417461853059085059819958060220257218705882557929843952675556987949500595639655528439052202",
			"2562656890570835296352376205216590519360952576353253013086344012422175466058176642832419040937235521572328705583208834436813588375562745525224328564354560731400723267162764903064018742843839822445601315505274421672289602485557719646504320106522113645676636456687468751723898215955665240524139836668877382766583475339565598073690853848639545227831264115164596396262772422415214665834769274554577301336288865874066248890243346947740610544045667761548206600923673948174739356732295677551749947395385332556227074205668024351973201205328576603362256016900712683688241615565934460363012498930253514800348031700419220337084",
			"99930612749194478118600296908042506926692739692582613847838885872895917361241984244333640602599304371167273758960563944657238964168847145890094008171141820109486094906362559078696139635411309492242182207416966460371128892064206746882387386758007030887894419440391740224416422066158827171353996674486071321206545809261340407889645960240945172487811751333275882478086448570781023042237458085693984643299076063991520418457060933571240987187072863856077205628245912721319981126278018820722990414471130623469536651100186034442507467323080021000888531326106364386993840733872502601891369558638525326664561154017828847672",
			"1472506745866574115259626864190224692004809117918934787166846528825105924609688841466108295556523488652122781674292297444292995560548157322502731129456683958121540244723666433763823118509403125487290004612301144261032505590317683722074031806725573843523098435377672727847103988767892621104751882955851738010346666992587376833949626244593568297587354693893573786831021202044714442955117877945963339672065587195524871759098746680962314094004915895821503042815585664974984481266334250938197696203435802781999374052416565642364570792202973860776695581419631252848636379297475201116274866616952484297948956399638888073235",
			"83104757259097189186950079721727937164047733672638603586459353944286466420114424207639146426104476267980570400699074675491296408383804082762803049161909665127121238239604267151066446049707844668847402565424593554013870425931673959230798027449894404276925930305256920765401340481960562820469278891866471990029",
			"71642128110699780750818062148317314680771514115102777629050114614166804537355399484072749627185009816049031339010525295431775519909278950430775769096405423548701628777704430192232627139339020595673609882501846596139297041913744680629890377085537497999944390724942020172896518996334279151212334736977022165629",
		),
	}

	membersTssPreParams := make(
		map[group.MemberIndex]*keygen.LocalPreParams,
		groupSize,
	)

	for memberID := group.MemberIndex(1); int(memberID) <= groupSize; memberID++ {
		// If there are predefined TSS pre-parameters for the given member -
		// take them. Otherwise, generate a new ones.
		tssPreParams, ok := predefinedMembersTssPreParams[memberID]
		if !ok {
			var err error
			tssPreParams, err = keygen.GeneratePreParams(1 * time.Minute)
			if err != nil {
				return nil, fmt.Errorf(
					"cannot generate TSS pre-parameters for member [%v]: [%v]",
					memberID,
					err,
				)
			}
		}

		membersTssPreParams[memberID] = tssPreParams
	}

	return membersTssPreParams, nil
}

func newTssPreParams(
	paillierSKN,
	paillierSKLambdaN,
	paillierSKPhiN,
	nTildei,
	h1i,
	h2i,
	alpha,
	beta,
	p,
	q string,
) *keygen.LocalPreParams {
	newBigInt := func(number string) *big.Int {
		bigIntNumber, _ := new(big.Int).SetString(number, 10)
		return bigIntNumber
	}

	return &keygen.LocalPreParams{
		PaillierSK: &paillier.PrivateKey{
			PublicKey: paillier.PublicKey{
				N: newBigInt(paillierSKN),
			},
			LambdaN: newBigInt(paillierSKLambdaN),
			PhiN:    newBigInt(paillierSKPhiN),
		},
		NTildei: newBigInt(nTildei),
		H1i:     newBigInt(h1i),
		H2i:     newBigInt(h2i),
		Alpha:   newBigInt(alpha),
		Beta:    newBigInt(beta),
		P:       newBigInt(p),
		Q:       newBigInt(q),
	}
}