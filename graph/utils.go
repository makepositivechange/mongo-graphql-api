package graph

import "github.com/google/uuid"

func Guidgenerator() string {
	return uuid.NewString()
}
