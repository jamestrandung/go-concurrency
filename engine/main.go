package main

import (
	"context"
	"fmt"

	"github.com/grab/async/engine/sample/server"
	"github.com/grab/async/engine/sample/service/miscellaneous"
	"github.com/grab/async/engine/sample/service/scaffolding"
)

func main() {
	server.Serve()

	p := scaffolding.NewPlan(
		miscellaneous.CostRequest{
			PointA: "Clementi",
			PointB: "Changi Airport",
		},
	)

	if err := p.Execute(context.Background()); err != nil {
		fmt.Println(err)
	}

	fmt.Println(p.GetTravelCost())
}
