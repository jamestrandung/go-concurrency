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
	// unregisters it from the pool of sold tickets.
	discardTicket(ctx context.Context)
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

func (b *ticketBooth) discardTicket(ctx context.Context) {
	v := ctx.Value(contextKey{})

	ticketID, ok := v.(string)
	if !ok {
		// No/invalid ticket was issued for the current goroutine
		return
	}

	arrived, ok := b.arrivedTickets[ticketID]
	if !ok {
		// This ticket has already been discarded
		return
	}

	delete(b.arrivedTickets, ticketID)
	if arrived {
		b.arrivedCount--
	}
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

	return b.arrivedCount == len(b.arrivedTickets)
}

type noOpTicketBooth struct{}

func (noOpTicketBooth) sellTicket(ctx context.Context) context.Context {
	return ctx
}

func (noOpTicketBooth) discardTicket(ctx context.Context) {
	// Do nothing
}

func (noOpTicketBooth) submitTicket(ctx context.Context) bool {
	return false
}
