package auth

import (
	"errors"
	"log"
	"time"
)

type FcCastID struct {
	FID  int
	Hash string
}

type FcFrameAction struct {
	Url         string
	ButtonIndex int
	InputText   string
	CastID      FcCastID
}

type FcFrameMessageData struct {
	Type            string
	Fid             int
	Timestamp       time.Time
	Network         string
	FrameActionBody FcFrameAction
}

type FcFrameMessage struct {
	Data            FcFrameMessageData
	Hash            string
	HashScheme      string
	Signature       string
	SignatureScheme string
	Signer          string
}

/**
 * ValidateFrameMessage validates a frame request message was signed by the claimed signer
 *
 * Unlike most frame/Farcaster validations, this one does not utilize
 * a Farcaster Hub node to validate the linkage of the signer to the Fid.
 * Thus...
 * - We do not need to make a Hub request or maintain a Hub
 * - We do not have validation of the claimed Fid
 *
 * @param message - The frame request message to validate
 * @return bool - True if the message was signed by the claimed signer
 * @return error - An error if the message was not signed by the claimed signer
 */
func ValidateFrameMessage(message FcFrameMessage) (bool, error) {
	/**
	 * TODO: Implement actual signature verification
	 * 0. Validate data
	 *   a. timestamp is within 5 minutes of current time
	 *   b. Network matches expectations
	 * 1. Validate hash matches hash scheme and message data
	 * 2. Validate signature matches signatureScheme, signer, and hash
	 * 3. Validate signature is correct for the message and signer
	 */

	// Log the incoming message for debugging
	log.Printf("Validating frame message: %+v\n", message)

	// 0a. Validate timestamp
	if message.Data.Timestamp.Add(5 * time.Minute).Before(time.Now()) {
		return false, errors.New("timestamp is too old")
	}

	// 0b. TODO: Validate network

	// 1. TODO: Validate hash matches hash scheme and message data

	// 2. TODO: Validate signature matches signatureScheme, signer, and hash

	// 3. TODO: Validate signature is correct for the message and signer

	return true, nil
}
