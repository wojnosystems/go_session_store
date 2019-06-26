package go_session_store

import (
	"context"
	"errors"
)

// SessionStorer defines an interface for the backend of a browser/app session storage system
type SessionStorer interface {
	// GenerateAndStore should use a SessionIdGenerator to create a new session and store the session so that it can be looked by the session Id later
	// @param ctx the context to use for timeouts, if required
	// @param userId the user that this session will represent
	// @param metaData is any data you wish to include when the session is looked up again
	// @return session identifier ideally created from a SessionIdGenerator
	// @return err errors encountered if saving the session or nil if no error occurred. If the generated session conflicted with an existing session, return ErrSessionCollision
	GenerateAndStore(ctx context.Context, userId string, metaData string) (session []byte, err error)

	// Get given a session created by GenerateAndStore., return the userId and metaData, if any. userId will be blank string if missing.
	// No ok sentinel value is used as it makes no sense to use GenerateAndStore without some sort of userId key
	// @param ctx the context to use for timeouts, if required
	// @param session the session that was created by GenerateAndStore.
	// @return userId the user that this session will represent. If no session exists, this will be an empty string
	// @return metaData is any data you wish to include when the session is looked up again. If no session exists, this will be an empty string
	// @return err the error encountered when looking up the session. This should NOT be a value representing no session
	Get(ctx context.Context, session []byte) (userId string, metaData string, err error)
}

type SessionIdGenerator interface {
	// Generate creates a new session Id. This was intended to use a crypt.Random source for generating sessions, but may not always be random
	// @return session the value to use to identify the session. As of 2019-06-25, this should be at least 128 bits (16 bytes) to ensure that it's hard to guess. You can always make this longer for more defense against guessing sessions
	// @return err any errors encountered when trying to generate a session Id.
	Generate() (session []byte, err error)
}

var ErrSessionCollision = errors.New("unable to store session, existing session ID already exists")

// New creates a new session and will attempt to retry it a maxGenerateAttempts time in case session generation is random. This is to prevent collisions from occurring
// if the SessionIdGenerator will never encounter a collision, this isn't needed, but most will NOT be this way ;). While it's unlikely to have a collision for 128 bit-wide sessions, it's not IMPOSSIBLE, but highly unlikely. In this case, you don't want to over-write the existing session.
// This method will try to set another session and will try to save it. If maxGenerateAttempts is exhausted, then it will return ErrSessionCollision.
// @param ctx the context to use for timing out network-based requests, for sessionStores that support it
// @param storer the session storage into which sessions are saved
// @param userId the user's identifier to use when looking up sessions to know which user (or whatever) the session is attached to
// @param metaData arbitrary data you wish to store with the userId.
// @param maxGenerateAttempts is the maximum number of times to try to generate and save the session Id before giving up. It will only retry if the error was due to a collision and not for other errors.
func New(ctx context.Context, storer SessionStorer, userId string, metaData string, maxGenerateAttempts int) (session []byte, err error) {
	for i := 0; i < maxGenerateAttempts; i++ {
		session, err = storer.GenerateAndStore(ctx, userId, metaData)
		if err == ErrSessionCollision {
			continue
		}
		return session, err
	}
	return []byte{}, ErrSessionCollision
}
