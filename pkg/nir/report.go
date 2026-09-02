package nir

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode/utf8"

	"nodora.org/nodora/pkg/core"
)

// caps how many invocations of the same lambda are rendered in a trace
const DefaultMaxLambdaFrames = 3

type ReportOptions struct {
	MaxLambdaFrames int
}

func (o ReportOptions) maxLambdaFrames() int {
	if o.MaxLambdaFrames <= 0 {
		return DefaultMaxLambdaFrames
	}
	return o.MaxLambdaFrames
}

type reportLine struct {
	indent string
	text   string
	value  string // empty means the line has no result column
	note   string
}

type Trace struct {
	RuleName  string
	Rule      *Rule
	Input     core.ValueMap
	Slots     []core.Value
	Emissions []EmittedSignal
	Root      *Event
}

func NewTrace(ruleName string, rule *Rule, input core.ValueMap, ctx *EvaluationContext, root *Event) *Trace {
	return &Trace{
		RuleName:  ruleName,
		Rule:      rule,
		Input:     input,
		Slots:     slices.Clone(ctx.Slots),
		Emissions: ctx.Emissions,
		Root:      root,
	}
}

func (t *Trace) Format(opts ReportOptions) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "--- trace %s ---\n\n", t.RuleName)
	writeInput(&sb, t.Input)

	sb.WriteString("steps\n")
	lines := make([]reportLine, 0, len(t.Root.Children)*3)
	for i, ev := range t.Root.Children {
		if ev.Kind != NodeOp {
			continue
		}
		collectOpLines(&lines, ev, fmt.Sprintf("%3d  ", i), opts)
	}
	writeLines(&sb, lines)

	// a failed run has no outputs and no meaningful slot state past the failure
	if failedOp(t.Root) {
		sb.WriteString("\n(no outputs produced)\n")
		writeSignals(&sb, t.Emissions)
		writeSlots(&sb, t.Slots)
		return sb.String()
	}

	writeOutputs(&sb, t.Rule, t.Slots)
	writeSignals(&sb, t.Emissions)
	writeSlots(&sb, t.Slots)

	return sb.String()
}

func failedOp(root *Event) bool {
	for _, ev := range root.Children {
		if ev.Kind == NodeOp && ev.Err != nil {
			return true
		}
	}
	return false
}

func writeInput(sb *strings.Builder, input core.ValueMap) {
	sb.WriteString("input\n")
	if len(input) == 0 {
		sb.WriteString("  (empty)\n\n")
		return
	}
	for _, k := range slices.Sorted(maps.Keys(input)) {
		fmt.Fprintf(sb, "  %s: %s\n", k, FormatValue(input[k]))
	}
	sb.WriteString("\n")
}

func dispLen(s string) int { return utf8.RuneCountInString(s) }

func writeLines(sb *strings.Builder, lines []reportLine) {
	width := 0
	for _, l := range lines {
		if l.value == "" {
			continue
		}
		if n := dispLen(l.indent) + dispLen(l.text); n > width {
			width = n
		}
	}

	pad := func(head string) string {
		if n := width - dispLen(head); n > 0 {
			return strings.Repeat(" ", n)
		}
		return ""
	}

	for _, l := range lines {
		head := l.indent + l.text
		if l.value == "" {
			sb.WriteString(head)
			if l.note != "" {
				sb.WriteString(pad(head))
				sb.WriteString("     ")
				sb.WriteString(l.note)
			}
			sb.WriteString("\n")
			continue
		}
		sb.WriteString(head)
		sb.WriteString(pad(head))
		sb.WriteString("  => ")
		sb.WriteString(l.value)
		if l.note != "" {
			sb.WriteString("   ")
			sb.WriteString(l.note)
		}
		sb.WriteString("\n")
	}
}

func collectOpLines(out *[]reportLine, ev *Event, prefix string, opts ReportOptions) {
	src := reportLine{indent: prefix, text: ev.Op.String()}

	value := ""
	if _, hasSlot := ev.Slot(); hasSlot {
		value = FormatValue(ev.Value)
	}
	sub := substitutedRHS(ev)

	// the substituted line continues the op, so it keeps the prefix's tree bars
	// and then pads its '=' to land under the source line's '='
	cont := continuationOf(prefix)
	if ev.Op.Out != nil {
		cont += strings.Repeat(" ", len(SlotRef(*ev.Op.Out))+1)
	}

	switch {
	case ev.Err != nil:
		// the out slot was never written; the failure note is the whole result
		src.note = errNote(ev)
		*out = append(*out, src)
	case ev.Op.Kind == OpEmit:
		*out = append(*out, src)
		*out = append(*out, reportLine{indent: cont, text: sub, note: emitNote(ev)})
	case sub == value:
		src.value = value
		*out = append(*out, src)
	default:
		*out = append(*out, src)
		*out = append(*out, reportLine{indent: cont, text: "= " + sub, value: value})
	}

	collectFrames(out, ev, cont+"  ", opts)
}

func substitutedRHS(ev *Event) string {
	return ev.Op.exprStringWith(func(i int) string {
		if i >= len(ev.Op.Args) {
			return "?"
		}
		return argSub(ev, ev.Op.Args[i])
	})
}

func argSub(parent *Event, raw RawExpr) string {
	for _, c := range parent.Children {
		if c.Kind == NodeExpr && c.Expr == raw.Expr {
			return substitute(c)
		}
	}
	return raw.String()
}

func substitutedArgs(parent *Event, args []RawExpr) string {
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = argSub(parent, a)
	}
	return strings.Join(parts, ", ")
}

func substitute(ev *Event) string {
	switch ex := ev.Expr.(type) {
	case *SignalExpr:
		out := ex.Name + "(" + substitutedArgs(ev, ex.Args) + ")"
		if ex.When != nil {
			out += " when " + argSub(ev, *ex.When)
		}
		return out
	case *CallExpr:
		return ex.Func.String() + "(" + substitutedArgs(ev, ex.Args) + ")"
	case *LambdaExpr:
		// the body is spelled out on the source line and expanded as frames
		return ex.Header()
	default:
		return FormatValue(ev.Value)
	}
}

func collectFrames(out *[]reportLine, ev *Event, prefix string, opts ReportOptions) {
	frames := gatherFrames(ev)
	if len(frames) == 0 {
		return
	}

	shown := len(frames)
	more := 0
	if cap := opts.maxLambdaFrames(); shown > cap {
		more, shown = shown-cap, cap
	}

	for i := 0; i < shown; i++ {
		frame := frames[i]
		branch, child := treeBranch(prefix, i == shown-1 && more == 0)
		*out = append(*out, reportLine{
			indent: branch,
			text:   frameBindings(frame),
			value:  FormatValue(frame.Value),
			note:   errNote(frame),
		})
		for _, op := range frameBodyOps(frame) {
			collectOpLines(out, op, child+"  ", opts)
		}
	}

	if more > 0 {
		branch, _ := treeBranch(prefix, true)
		*out = append(*out, reportLine{
			indent: branch,
			text:   fmt.Sprintf("... %d more invocation(s)", more),
		})
	}
}

func gatherFrames(ev *Event) []*Event {
	var frames []*Event
	for _, c := range ev.Children {
		switch c.Kind {
		case NodeLambda:
			frames = append(frames, c)
		case NodeExpr:
			switch c.Expr.(type) {
			case *ImmExpr, *SymExpr, *LambdaExpr:
			default:
				frames = append(frames, gatherFrames(c)...)
			}
		}
	}
	return frames
}

func treeBranch(prefix string, last bool) (branch, child string) {
	if last {
		return prefix + "└─ ", prefix + "   "
	}
	return prefix + "├─ ", prefix + "│  "
}

func continuationOf(prefix string) string {
	var b strings.Builder
	for _, r := range prefix {
		if r == '│' {
			b.WriteRune(r)
		} else {
			b.WriteByte(' ')
		}
	}
	return b.String()
}

func frameBindings(frame *Event) string {
	le, ok := frame.Lambda()
	parts := make([]string, len(frame.Args))
	for i, a := range frame.Args {
		name := "?"
		if ok && i < len(le.Params) {
			name = SlotRef(le.Params[i].SymIndex)
		}
		parts[i] = name + " <= " + FormatValue(a)
	}
	return strings.Join(parts, ", ")
}

func frameBodyOps(frame *Event) []*Event {
	ops := make([]*Event, 0, len(frame.Children))
	for _, inner := range frame.Children {
		if inner.Kind == NodeOp {
			ops = append(ops, inner)
		}
	}
	return ops
}

func errNote(ev *Event) string {
	if ev.Err != nil {
		return "error: " + ev.Err.Error()
	}
	return ""
}

func emitNote(ev *Event) string {
	if len(ev.Children) == 0 {
		return ""
	}
	signal, ok := ev.Children[0].Expr.(*SignalExpr)
	if !ok {
		return ""
	}
	if signal.When == nil {
		return "emitted"
	}
	guard := ev.Children[0].Children
	if len(guard) == 0 {
		return ""
	}
	if b, isBool := guard[0].Value.Raw.(bool); isBool && b {
		return "emitted"
	}
	return "not emitted"
}

func writeOutputs(sb *strings.Builder, rule *Rule, slots []core.Value) {
	if rule == nil || len(rule.Outputs) == 0 {
		return
	}
	sb.WriteString("\noutputs\n")
	for _, name := range slices.Sorted(maps.Keys(rule.Outputs)) {
		val := slotValue(slots, rule.Outputs[name].Sym)
		line := fmt.Sprintf("  %s = %s => %s", name, SlotRef(rule.Outputs[name].Sym), FormatValue(val))
		if val.IsUndefined() {
			line += "   (omitted from result)"
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}
}

func writeSlots(sb *strings.Builder, slots []core.Value) {
	if len(slots) <= 1 {
		return
	}
	width := 0
	for i := 1; i < len(slots); i++ {
		if n := len(SlotRef(i)); n > width {
			width = n
		}
	}
	sb.WriteString("\nslotdump\n")
	for i := 1; i < len(slots); i++ {
		fmt.Fprintf(sb, "  %-*s => %s\n", width, SlotRef(i), FormatValue(slots[i]))
	}
}

func writeSignals(sb *strings.Builder, emissions []EmittedSignal) {
	if len(emissions) == 0 {
		return
	}
	sb.WriteString("\nsignals\n")
	for _, e := range emissions {
		args := make([]string, len(e.Args))
		for i, a := range e.Args {
			args[i] = FormatValue(core.V(a))
		}
		fmt.Fprintf(sb, "  %s(%s)\n", e.Name, strings.Join(args, ", "))
	}
}

func slotValue(slots []core.Value, slot int) core.Value {
	if slot < 0 || slot >= len(slots) {
		return core.U()
	}
	return slots[slot]
}
