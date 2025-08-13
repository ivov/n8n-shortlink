package errors

import stdErrors "errors"

var (
	// ErrShortlinkNotFound is returned when a shortlink is not found.
	ErrShortlinkNotFound = stdErrors.New("shortlink not found")

	// ErrSlugTaken is returned when a slug is already taken.
	ErrSlugTaken = stdErrors.New("custom slug is already taken")

	// ErrSlugMisformatted is returned when a slug is invalid.
	ErrSlugMisformatted = stdErrors.New("custom slug is misformatted - must contain only A-Z, a-z, 0-9, -, _")

	// ErrSlugTooShort is returned when a slug is too short.
	ErrSlugTooShort = stdErrors.New("custom slug is too short - min 4 chars")

	// ErrSlugTooLong is returned when a slug is too long.
	ErrSlugTooLong = stdErrors.New("custom slug is too long - max 512 chars")

	// ErrSlugReserved is returned when a slug is reserved for internal use.
	ErrSlugReserved = stdErrors.New("custom slug is reserved for internal use")

	// ErrAuthHeaderMissing is returned when the authorization header is missing.
	ErrAuthHeaderMissing = stdErrors.New("authorization header is missing")

	// ErrAuthHeaderMalformed is returned when the Authorization header is invalid.
	ErrAuthHeaderMalformed = stdErrors.New("authorization header is malformed")

	// ErrPasswordInvalid is returned when the password is invalid.
	ErrPasswordInvalid = stdErrors.New("password is invalid")

	// ErrKindUnsupported is returned when the shortlink kind is unsupported.
	ErrKindUnsupported = stdErrors.New("shortlink kind is unsupported - neither \"url\" nor \"workflow\"")

	// ErrContentMalformed is returned when the content is malformed.
	ErrContentMalformed = stdErrors.New("content is malformed - neither URL nor JSON")

	// ErrPasswordTooShort is returned when the password is too short.
	ErrPasswordTooShort = stdErrors.New("password is too short - must be at least 8 chars")

	// ErrPayloadTooLarge is returned when the payload is too large.
	ErrPayloadTooLarge = stdErrors.New("payload is too large - max size is 5 MB")

	// ErrContentBlocked is returned when content contains suspicious patterns.
	ErrContentBlocked = stdErrors.New("content blocked - suspicious pattern detected")
)

// ToCode maps errors to error codes.
var ToCode = map[error]string{
	ErrShortlinkNotFound:   "SHORTLINK_NOT_FOUND",
	ErrKindUnsupported:     "KIND_UNSUPPORTED",
	ErrSlugTaken:           "SLUG_TAKEN",
	ErrSlugMisformatted:    "SLUG_MISFORMATTED",
	ErrSlugTooShort:        "SLUG_TOO_SHORT",
	ErrSlugTooLong:         "SLUG_TOO_LONG",
	ErrSlugReserved:        "SLUG_RESERVED",
	ErrAuthHeaderMissing:   "AUTHORIZATION_HEADER_MISSING",
	ErrAuthHeaderMalformed: "AUTHORIZATION_HEADER_MALFORMED",
	ErrContentMalformed:    "CONTENT_MALFORMED",
	ErrPasswordTooShort:    "PASSWORD_TOO_SHORT",
	ErrPayloadTooLarge:     "PAYLOAD_TOO_LARGE",
	ErrPasswordInvalid:     "PASSWORD_INVALID",
	ErrContentBlocked:      "CONTENT_BLOCKED",
}
