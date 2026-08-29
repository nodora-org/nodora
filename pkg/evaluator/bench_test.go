package evaluator_test

import (
	"testing"

	"nodora.org/nodora/pkg/compiler"
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/evaluator"
	_ "nodora.org/nodora/pkg/registry/all"
)

const benchRuleset = `
signal HighValue(total)

rule Bench {
    subtotal   = input.order.subtotal
    shipping   = input.order.shipping
    total      = subtotal + shipping
    discount   = input.order.discount
    discounted = total * (1 - discount)

    tierRate = match input.user.tier {
        "gold"   => 0.2,
        "silver" => 0.1,
        _        => 0.0
    }

    tags      = ["new", "priority", "vip"]
    expensive = arrays::filter(input.order.items, |it| it.price > 10)

    out finalTotal   = discounted
    out over         = discounted > 100
    out tier         = tierRate
    out isVip        = "vip" in tags
    out numExpensive = len(expensive)
    out absAdj       = math::abs(input.order.adjustment)

    emit HighValue(discounted) when discounted > 500
}
`

func benchInput() core.ValueMap {
	return core.ValueMap{
		"order": core.V(map[string]any{
			"subtotal":   120.0,
			"shipping":   15.0,
			"discount":   0.1,
			"adjustment": -7.5,
			"items": []any{
				map[string]any{"price": 5.0},
				map[string]any{"price": 25.0},
				map[string]any{"price": 12.0},
				map[string]any{"price": 3.0},
			},
		}),
		"user": core.V(map[string]any{"tier": "gold"}),
	}
}

func BenchmarkEvaluateRule(b *testing.B) {
	ruleset, err := compiler.NewCompiler().Compile(benchRuleset)
	if err != nil {
		b.Fatalf("compile: %v", err)
	}
	ev := evaluator.NewEvaluator(ruleset)
	input := benchInput()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ev.EvaluateRule("Bench", input); err != nil {
			b.Fatalf("evaluate: %v", err)
		}
	}
}
