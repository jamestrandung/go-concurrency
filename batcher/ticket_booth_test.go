package batcher

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTicketBooth_SellTicket(t *testing.T) {
	b := newTicketBooth()
	assert.Equal(t, 0, len(b.arrivedTickets))
	assert.Equal(t, 0, b.arrivedCount)

	ctxWithTicket := b.sellTicket(context.Background())

	assert.Equal(t, 1, len(b.arrivedTickets))
	assert.Equal(t, 0, b.arrivedCount)

	v := ctxWithTicket.Value(contextKey{})
	assert.NotNil(t, v)

	ticketID, ok := v.(string)
	if assert.True(t, ok) {
		arrived, ok := b.arrivedTickets[ticketID]
		assert.True(t, ok)
		assert.False(t, arrived)
	}
}

func TestTicketBooth_DiscardTicket(t *testing.T) {
	b := newTicketBooth()
	assert.Equal(t, 0, len(b.arrivedTickets))
	assert.Equal(t, 0, b.arrivedCount)

	ctxWithTicket := b.sellTicket(context.Background())

	assert.Equal(t, 1, len(b.arrivedTickets))
	assert.Equal(t, 0, b.arrivedCount)

	// Discard a ticket that has never arrived
	b.discardTicket(ctxWithTicket)

	assert.Equal(t, 0, len(b.arrivedTickets))
	assert.Equal(t, 0, b.arrivedCount)

	ctxWithTicket = b.sellTicket(context.Background())

	assert.Equal(t, 1, len(b.arrivedTickets))
	assert.Equal(t, 0, b.arrivedCount)

	b.submitTicket(ctxWithTicket)

	assert.Equal(t, 1, len(b.arrivedTickets))
	assert.Equal(t, 1, b.arrivedCount)

	// Discard a ticket that has already arrived
	b.discardTicket(ctxWithTicket)

	assert.Equal(t, 0, len(b.arrivedTickets))
	assert.Equal(t, 0, b.arrivedCount)
}

func TestTicketBooth_SubmitTicket(t *testing.T) {
	b := newTicketBooth()

	ctx := context.Background()

	// Ctx with no ticket
	assert.False(t, b.submitTicket(ctx))

	ctxWithTicket := b.sellTicket(ctx)

	// Ctx with ticket that has never arrived
	assert.True(t, b.submitTicket(ctxWithTicket))
	assert.Equal(t, 1, b.arrivedCount)

	// Ctx with ticket that has already arrived
	assert.False(t, b.submitTicket(ctxWithTicket))
	assert.Equal(t, 1, b.arrivedCount)

	ctxWithDiscardedTicket := b.sellTicket(ctx)
	b.discardTicket(ctxWithDiscardedTicket)

	// Ctx with ticket that has already been discarded
	assert.False(t, b.submitTicket(ctxWithDiscardedTicket))
}
