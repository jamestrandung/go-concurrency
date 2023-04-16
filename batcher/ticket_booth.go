package batcher

import (
	"context"

	"github.com/twinj/uuid"
)

// iTicketBooth is a synchronous ticket generator that can track how many
// ticket owners have arrived.
type iTicketBooth interface {
	// sellTicket issues a new ticket and embeds it in the given context.
	sellTicket(ctx context.Context) context.Context
	// discardTicket finds an issued ticket in the given context and if found,
	// unregisters it from the pool of sold tickets and then returns whether
	// all ticket owners have arrived.
	discardTicket(ctx context.Context) bool
	// submitTicket finds an issued ticket in the given context and if found,
	// marks it as arrived and returns whether all ticket owners have arrived.
	submitTicket(ctx context.Context) bool
}

type ticketBooth struct {
	arrivedTickets map[string]bool
	arrivedCount   int
}

func newTicketBooth() *ticketBooth {
	return &ticketBooth{
		arrivedTickets: make(map[string]bool),
	}
}

type contextKey struct{}

func (b *ticketBooth) sellTicket(ctx context.Context) context.Context {
	ticketID := uuid.NewV4().String()
	b.arrivedTickets[ticketID] = false

	return context.WithValue(ctx, contextKey{}, ticketID)
}

func (b *ticketBooth) discardTicket(ctx context.Context) bool {
	v := ctx.Value(contextKey{})

	ticketID, ok := v.(string)
	if !ok {
		// No/invalid ticket was issued for the current goroutine
		return false
	}

	arrived, ok := b.arrivedTickets[ticketID]
	if !ok {
		// This ticket has already been discarded
		return false
	}

	delete(b.arrivedTickets, ticketID)
	if arrived {
		b.arrivedCount--
	}

	return b.resetIfAllTicketsHaveArrived()
}

func (b *ticketBooth) submitTicket(ctx context.Context) bool {
	v := ctx.Value(contextKey{})

	ticketID, ok := v.(string)
	if !ok {
		// No/invalid ticket was issued for the current goroutine
		return false
	}

	arrived, ok := b.arrivedTickets[ticketID]
	if !ok || arrived {
		// This ticket has already been discarded or submitted before
		return false
	}

	b.arrivedTickets[ticketID] = true
	b.arrivedCount++

	return b.resetIfAllTicketsHaveArrived()
}

func (b *ticketBooth) resetIfAllTicketsHaveArrived() bool {
	// Might be 0 when a ticket is discarded
	if len(b.arrivedTickets) == 0 {
		return false
	}

	haveAllArrived := b.arrivedCount == len(b.arrivedTickets)
	if haveAllArrived {
		// Reset so that clients can repeat if necessary
		b.arrivedCount = 0
		for k, _ := range b.arrivedTickets {
			b.arrivedTickets[k] = false
		}
	}

	return haveAllArrived
}

type noOpTicketBooth struct{}

func (noOpTicketBooth) sellTicket(ctx context.Context) context.Context {
	return ctx
}

func (noOpTicketBooth) discardTicket(ctx context.Context) bool {
	return false
}

func (noOpTicketBooth) submitTicket(ctx context.Context) bool {
	return false
}
