package utils

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ErrInvalidMongoID is a custom error for invalid MongoDB ObjectIDs.
var ErrInvalidMongoID = errors.New("the provided value is not a valid MongoDB ObjectID")

// ValidateAndParseMongoID validates a string as a MongoDB ObjectID and converts it to primitive.ObjectID.
// Returns the parsed ObjectID or an error if the ID is invalid.
func ValidateAndParseMongoID(id string) (primitive.ObjectID, error) {
	if id == "" {
		return primitive.ObjectID{}, ErrInvalidMongoID
	}

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.ObjectID{}, ErrInvalidMongoID
	}

	return oid, nil
}

// For processing multiple IDs
func ValidateAndParseMongoIDs(ids []string) ([]primitive.ObjectID, error) {
	result := make([]primitive.ObjectID, len(ids))
	for i, id := range ids {
		oid, err := ValidateAndParseMongoID(id)
		if err != nil {
			return nil, fmt.Errorf("invalid ID at index %d: %w", i, err)
		}
		result[i] = oid
	}
	return result, nil
}
