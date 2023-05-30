package batcher

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestTicketBooth_SellTicket(t *testing.T) {
    batcherID := "batcherID"

    b := newTicketBooth()
    assert.Equal(t, 0, len(b.arrivedTickets))
    assert.Equal(t, 0, b.arrivedCount)

    ctxWithTicket := b.sellTicket(context.Background(), batcherID)

    assert.Equal(t, 1, len(b.arrivedTickets))
    assert.Equal(t, 0, b.arrivedCount)

    v := ctxWithTicket.Value(contextKey{batcherID})
    assert.NotNil(t, v)

    ticketID, ok := v.(string)
    if assert.True(t, ok) {
        arrived, ok := b.arrivedTickets[ticketID]
        assert.True(t, ok)
        assert.False(t, arrived)
    }
}

func TestTicketBooth_DiscardTicket(t *testing.T) {
    batcherID := "batcherID"

    b := newTicketBooth()
    assert.Equal(t, 0, len(b.arrivedTickets))
    assert.Equal(t, 0, b.arrivedCount)

    ctxWithTicket := b.sellTicket(context.Background(), batcherID)
    ctxWithTicket2 := b.sellTicket(context.Background(), batcherID)

    assert.Equal(t, 2, len(b.arrivedTickets))
    assert.Equal(t, 0, b.arrivedCount)

    // Discard a ticket that has never arrived
    haveAllArrived := b.discardTicket(ctxWithTicket, batcherID)

    assert.False(t, haveAllArrived)
    assert.Equal(t, 1, len(b.arrivedTickets))
    assert.Equal(t, 0, b.arrivedCount)

    ctxWithTicket = b.sellTicket(context.Background(), batcherID)

    assert.Equal(t, 2, len(b.arrivedTickets))
    assert.Equal(t, 0, b.arrivedCount)

    b.submitTicket(ctxWithTicket, batcherID)

    assert.Equal(t, 2, len(b.arrivedTickets))
    assert.Equal(t, 1, b.arrivedCount)

    // Discard a ticket that has already arrived
    // leaving 1 ticket
    haveAllArrived = b.discardTicket(ctxWithTicket, batcherID)

    assert.False(t, haveAllArrived)
    assert.Equal(t, 1, len(b.arrivedTickets))
    assert.Equal(t, 0, b.arrivedCount)

    // Discard the other ticket that has never arrived
    // leaving no tickets left
    haveAllArrived = b.discardTicket(ctxWithTicket2, batcherID)

    assert.False(t, haveAllArrived)
    assert.Equal(t, 0, len(b.arrivedTickets))
    assert.Equal(t, 0, b.arrivedCount)

    // Discard the other ticket that has never arrived
    // leaving 1 ticket that has already arrived
    ctxWithTicket = b.sellTicket(context.Background(), batcherID)
    ctxWithTicket2 = b.sellTicket(context.Background(), batcherID)
    b.submitTicket(ctxWithTicket, batcherID)

    haveAllArrived = b.discardTicket(ctxWithTicket2, batcherID)

    assert.True(t, haveAllArrived)
    assert.Equal(t, 1, len(b.arrivedTickets))
    assert.Equal(t, 0, b.arrivedCount)
}

func TestTicketBooth_SubmitTicket(t *testing.T) {
    batcherID := "batcherID"

    b := newTicketBooth()

    ctx := context.Background()

    // Ctx with no ticket
    assert.False(t, b.submitTicket(ctx, batcherID))

    ctxWithTicket1 := b.sellTicket(ctx, batcherID)
    ctxWithTicket2 := b.sellTicket(ctx, batcherID)

    // Ctx with ticket that has never arrived
    assert.False(t, b.submitTicket(ctxWithTicket1, batcherID))
    assert.Equal(t, 1, b.arrivedCount)

    // Ctx with ticket that has already arrived
    assert.False(t, b.submitTicket(ctxWithTicket1, batcherID))
    assert.Equal(t, 1, b.arrivedCount)

    ctxWithDiscardedTicket := b.sellTicket(ctx, batcherID)
    b.discardTicket(ctxWithDiscardedTicket, batcherID)

    // Ctx with ticket that has already been discarded
    assert.False(t, b.submitTicket(ctxWithDiscardedTicket, batcherID))

    // Ctx with ticket that has never arrived
    assert.True(t, b.submitTicket(ctxWithTicket2, batcherID))
    // After all clients have arrived, should reset
    assert.Equal(t, 0, b.arrivedCount)
}
