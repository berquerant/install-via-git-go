package strategy_test

import (
	"berquerant/install-via-git-go/strategy"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoopRunner(t *testing.T) {
	err := strategy.NewNoopRunner().Run(context.TODO())
	assert.ErrorIs(t, err, strategy.ErrNoopStrategy)
}

func TestUnknownRunner(t *testing.T) {
	err := strategy.NewUnknownRunner().Run(context.TODO())
	assert.ErrorIs(t, err, strategy.ErrUnknownStrategy)
}
