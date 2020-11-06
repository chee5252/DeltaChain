// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package stack analyzes stack dump of Go processes and simplifies it.
//
// It is mostly useful on servers will large number of identical goroutines,
// making the crash dump harder to read than strictly necesary.
package stack

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const lockedToThread = "locked to thread"

var (
	// TODO(maruel): Handle corrupted stack cases:
	// - missed stack barrier
	// - found next stack barrier at 0x123; expected
	// - runtime: unexpected return pc for FUNC_NAME called from 0x123

	reRoutineHeader = regexp.MustCompile("^goroutine (\\d+) \\[([^\\]]+)\\]\\:\n$")
	reMinutes       = regexp.MustCompile("^(\\d+) minutes$")
	reUnavail       = regexp.MustCompile("^(?:\t| +)goroutine running on other thread; stack unavailable")
	// See gentraceback() in src/runtime/traceback.go for more information.
	// - Sometimes the source file comes up as "<autogenerated>". It is the
	//   compiler than generated these, not the runtime.
	// - The tab may be replaced with spaces when a user copy-paste it, handle
	//   this transparently.
	// - "runtime.gopanic" is explicitly replaced with "panic" by gentraceback().
	// - The +0x123 byte offset is printed when frame.pc > _func.entry. _func is
	//   generated by the linker.
	// - The +0x123 byte offset is not included with generated code, e.g. unnamed
	//   functions "func·006()" which is generally go func() { ... }()
	//   statements. Since the _func is generated at runtime, it's probably why
	//   _func.entry is not set.
	// - C calls may have fp=0x123 sp=0x123 appended. I think it normally happens
	//   when a signal is not correctly handled. It is printed with m.throwing>0.
	//   These are discarded.
	// - For cgo, the source file may be "??".
	reFile = regexp.MustCompile("^(?:\t| +)(\\?\\?|\\<autogenerated\\>|.+\\.(?:c|go|s))\\:(\\d+)(?:| \\+0x[0-9a-f]+)(?:| fp=0x[0-9a-f]+ sp=0x[0-9a-f]+)\n$")
	// Sadly, it doesn't note the goroutine number so we could cascade them per
	// parenthood.
	reCreated = regexp.MustCompile("^created by (.+)\n$")
	reFunc    = regexp.MustCompile("^(.+)\\((.*)\\)\n$")
	reElided  = regexp.MustCompile("^\\.\\.\\.additional frames elided\\.\\.\\.\n$")
	// Include frequent GOROOT value on Windows, distro provided and user
	// installed path. This simplifies the user's life when processing a trace
	// generated on another VM.
	// TODO(maruel): Guess the path automatically via traces containing the
	// 'runtime' package, which is very frequent. This would be "less bad" than
	// throwing up random values at the parser.
	goroots = []string{runtime.GOROOT(), "c:/go", "/usr/lib/go", "/usr/local/go"}
)

// Similarity is the level at which two call lines arguments must match to be
// considered similar enough to coalesce them.
type Similarity int

const (
	// ExactFlags requires same bits (e.g. Locked).
	ExactFlags Similarity = iota
	// ExactLines requests the exact same arguments on the call line.
	ExactLines
	// AnyPointer considers different pointers a similar call line.
	AnyPointer
	// AnyValue accepts any value as similar call line.
	AnyValue
)

// Function is a function call.
//
// Go stack traces print a mangled function call, this wrapper unmangle the
// string before printing and adds other filtering mdchods.
type Function struct {
	Raw string
}

// String is the fully qualified function name.
//
// Sadly Go is a bit confused when the package name doesn't match the directory
// containing the source file and will use the directory name instead of the
// real package name.
func (f Function) String() string {
	s, _ := url.QueryUnescape(f.Raw)
	return s
}

// Name is the naked function name.
func (f Function) Name() string {
	parts := strings.SplitN(filepath.Base(f.Raw), ".", 2)
	if len(parts) == 1 {
		return parts[0]
	}
	return parts[1]
}

// PkgName is the package name for this function reference.
func (f Function) PkgName() string {
	parts := strings.SplitN(filepath.Base(f.Raw), ".", 2)
	if len(parts) == 1 {
		return ""
	}
	s, _ := url.QueryUnescape(parts[0])
	return s
}

// PkgDotName returns "<package>.<func>" format.
func (f Function) PkgDotName() string {
	parts := strings.SplitN(filepath.Base(f.Raw), ".", 2)
	s, _ := url.QueryUnescape(parts[0])
	if len(parts) == 1 {
		return parts[0]
	}
	if s != "" || parts[1] != "" {
		return s + "." + parts[1]
	}
	return ""
}

// IsExported returns true if the function is exported.
func (f Function) IsExported() bool {
	name := f.Name()
	parts := strings.Split(name, ".")
	r, _ := utf8.DecodeRuneInString(parts[len(parts)-1])
	if unicode.ToUpper(r) == r {
		return true
	}
	return f.PkgName() == "main" && name == "main"
}

// Arg is an argument on a Call.
type Arg struct {
	Value uint64 // Value is the raw value as found in the stack trace
	Name  string // Name is a pseudo name given to the argument
}

// IsPtr returns true if we guess it's a pointer. It's only a guess, it can be
// easily be confused by a bitmask.
func (a *Arg) IsPtr() bool {
	// Assumes all pointers are above 16Mb and positive.
	return a.Value > 16*1024*1024 && a.Value < math.MaxInt64
}

func (a Arg) String() string {
	if a.Name != "" {
		return a.Name
	}
	if a.Value == 0 {
		return "0"
	}
	return fmt.Sprintf("0x%x", a.Value)
}

// Args is a series of function call arguments.
type Args struct {
	Values    []Arg    // Values is the arguments as shown on the stack trace. They are mangled via simplification.
	Processed []string // Processed is the arguments generated from processing the source files. It can have a length lower than Values.
	Elided    bool     // If set, it means there was a trailing ", ..."
}

func (a Args) String() string {
	var v []string
	if len(a.Processed) != 0 {
		v = make([]string, 0, len(a.Processed))
		for _, item := range a.Processed {
			v = append(v, item)
		}
	} else {
		v = make([]string, 0, len(a.Values))
		for _, item := range a.Values {
			v = append(v, item.String())
		}
	}
	if a.Elided {
		v = append(v, "...")
	}
	return strings.Join(v, ", ")
}

// Equal returns true only if both arguments are exactly equal.
func (a *Args) Equal(r *Args) bool {
	if a.Elided != r.Elided || len(a.Values) != len(r.Values) {
		return false
	}
	for i, l := range a.Values {
		if l != r.Values[i] {
			return false
		}
	}
	return true
}

// Similar returns true if the two Args are equal or almost but not quite
// equal.
func (a *Args) Similar(r *Args, similar Similarity) bool {
	if a.Elided != r.Elided || len(a.Values) != len(r.Values) {
		return false
	}
	if similar == AnyValue {
		return true
	}
	for i, l := range a.Values {
		switch similar {
		case ExactFlags, ExactLines:
			if l != r.Values[i] {
				return false
			}
		default:
			if l.IsPtr() != r.Values[i].IsPtr() || (!l.IsPtr() && l != r.Values[i]) {
				return false
			}
		}
	}
	return true
}

// Merge merges two similar Args, zapping out differences.
func (a *Args) Merge(r *Args) Args {
	out := Args{
		Values: make([]Arg, len(a.Values)),
		Elided: a.Elided,
	}
	for i, l := range a.Values {
		if l != r.Values[i] {
			out.Values[i].Name = "*"
			out.Values[i].Value = l.Value
		} else {
			out.Values[i] = l
		}
	}
	return out
}

// Call is an item in the stack trace.
type Call struct {
	SourcePath string   // Full path name of the source file
	Line       int      // Line number
	Func       Function // Fully qualified function name (encoded).
	Args       Args     // Call arguments
}

// Equal returns true only if both calls are exactly equal.
func (c *Call) Equal(r *Call) bool {
	return c.SourcePath == r.SourcePath && c.Line == r.Line && c.Func == r.Func && c.Args.Equal(&r.Args)
}

// Similar returns true if the two Call are equal or almost but not quite
// equal.
func (c *Call) Similar(r *Call, similar Similarity) bool {
	return c.SourcePath == r.SourcePath && c.Line == r.Line && c.Func == r.Func && c.Args.Similar(&r.Args, similar)
}

// Merge merges two similar Call, zapping out differences.
func (c *Call) Merge(r *Call) Call {
	return Call{
		SourcePath: c.SourcePath,
		Line:       c.Line,
		Func:       c.Func,
		Args:       c.Args.Merge(&r.Args),
	}
}

// SourceName returns the base file name of the source file.
func (c *Call) SourceName() string {
	return filepath.Base(c.SourcePath)
}

// SourceLine returns "source.go:line", including only the base file name.
func (c *Call) SourceLine() string {
	return fmt.Sprintf("%s:%d", c.SourceName(), c.Line)
}

// FullSourceLine returns "/path/to/source.go:line".
func (c *Call) FullSourceLine() string {
	return fmt.Sprintf("%s:%d", c.SourcePath, c.Line)
}

// PkgSource is one directory plus the file name of the source file.
func (c *Call) PkgSource() string {
	return filepath.Join(filepath.Base(filepath.Dir(c.SourcePath)), c.SourceName())
}

const testMainSource = "_test" + string(os.PathSeparator) + "_testmain.go"

// IsStdlib returns true if it is a Go standard library function. This includes
// the 'go test' generated main executable.
func (c *Call) IsStdlib() bool {
	for _, goroot := range goroots {
		if strings.HasPrefix(c.SourcePath, goroot) {
			return true
		}
	}
	// Consider _test/_testmain.go as stdlib since it's injected by "go test".
	return c.PkgSource() == testMainSource
}

// IsPkgMain returns true if it is in the main package.
func (c *Call) IsPkgMain() bool {
	return c.Func.PkgName() == "main"
}

// Stack is a call stack.
type Stack struct {
	Calls  []Call // Call stack. First is original function, last is leaf function.
	Elided bool   // Happens when there's >100 items in Stack, currently hardcoded in package runtime.
}

// Equal returns true on if both call stacks are exactly equal.
func (s *Stack) Equal(r *Stack) bool {
	if len(s.Calls) != len(r.Calls) || s.Elided != r.Elided {
		return false
	}
	for i := range s.Calls {
		if !s.Calls[i].Equal(&r.Calls[i]) {
			return false
		}
	}
	return true
}

// Similar returns true if the two Stack are equal or almost but not quite
// equal.
func (s *Stack) Similar(r *Stack, similar Similarity) bool {
	if len(s.Calls) != len(r.Calls) || s.Elided != r.Elided {
		return false
	}
	for i := range s.Calls {
		if !s.Calls[i].Similar(&r.Calls[i], similar) {
			return false
		}
	}
	return true
}

// Merge merges two similar Stack, zapping out differences.
func (s *Stack) Merge(r *Stack) *Stack {
	// Assumes similar stacks have the same length.
	out := &Stack{
		Calls:  make([]Call, len(s.Calls)),
		Elided: s.Elided,
	}
	for i := range s.Calls {
		out.Calls[i] = s.Calls[i].Merge(&r.Calls[i])
	}
	return out
}

// Less compares two Stack, where the ones that are less are more
// important, so they come up front. A Stack with more private functions is
// 'less' so it is at the top. Inversely, a Stack with only public
// functions is 'more' so it is at the bottom.
func (s *Stack) Less(r *Stack) bool {
	lStdlib := 0
	lPrivate := 0
	for _, c := range s.Calls {
		if c.IsStdlib() {
			lStdlib++
		} else {
			lPrivate++
		}
	}
	rStdlib := 0
	rPrivate := 0
	for _, s := range r.Calls {
		if s.IsStdlib() {
			rStdlib++
		} else {
			rPrivate++
		}
	}
	if lPrivate > rPrivate {
		return true
	}
	if lPrivate < rPrivate {
		return false
	}
	if lStdlib > rStdlib {
		return false
	}
	if lStdlib < rStdlib {
		return true
	}

	// Stack lengths are the same.
	for x := range s.Calls {
		if s.Calls[x].Func.Raw < r.Calls[x].Func.Raw {
			return true
		}
		if s.Calls[x].Func.Raw > r.Calls[x].Func.Raw {
			return true
		}
		if s.Calls[x].PkgSource() < r.Calls[x].PkgSource() {
			return true
		}
		if s.Calls[x].PkgSource() > r.Calls[x].PkgSource() {
			return true
		}
		if s.Calls[x].Line < r.Calls[x].Line {
			return true
		}
		if s.Calls[x].Line > r.Calls[x].Line {
			return true
		}
	}
	return false
}

// Signature represents the signature of one or multiple goroutines.
//
// It is effectively the stack trace plus the goroutine internal bits, like
// it's state, if it is thread locked, which call site created this goroutine,
// etc.
type Signature struct {
	// Use git grep 'gopark(|unlock)\(' to find them all plus everything listed
	// in runtime/traceback.go. Valid values includes:
	//     - chan send, chan receive, select
	//     - finalizer wait, mark wait (idle),
	//     - Concurrent GC wait, GC sweep wait, force gc (idle)
	//     - IO wait, panicwait
	//     - semacquire, semarelease
	//     - sleep, timer goroutine (idle)
	//     - trace reader (blocked)
	// Stuck cases:
	//     - chan send (nil chan), chan receive (nil chan), select (no cases)
	// Runnable states:
	//    - idle, runnable, running, syscall, waiting, dead, enqueue, copystack,
	// Scan states:
	//    - scan, scanrunnable, scanrunning, scansyscall, scanwaiting, scandead,
	//      scanenqueue
	State     string
	CreatedBy Call // Which other goroutine which created this one.
	SleepMin  int  // Wait time in minutes, if applicable.
	SleepMax  int  // Wait time in minutes, if applicable.
	Stack     Stack
	Locked    bool // Locked to an OS thread.
}

// Equal returns true only if both signatures are exactly equal.
func (s *Signature) Equal(r *Signature) bool {
	if s.State != r.State || !s.CreatedBy.Equal(&r.CreatedBy) || s.Locked != r.Locked || s.SleepMin != r.SleepMin || s.SleepMax != r.SleepMax {
		return false
	}
	return s.Stack.Equal(&r.Stack)
}

// Similar returns true if the two Signature are equal or almost but not quite
// equal.
func (s *Signature) Similar(r *Signature, similar Similarity) bool {
	if s.State != r.State || !s.CreatedBy.Similar(&r.CreatedBy, similar) {
		return false
	}
	if similar == ExactFlags && s.Locked != r.Locked {
		return false
	}
	return s.Stack.Similar(&r.Stack, similar)
}

// Merge merges two similar Signature, zapping out differences.
func (s *Signature) Merge(r *Signature) *Signature {
	min := s.SleepMin
	if r.SleepMin < min {
		min = r.SleepMin
	}
	max := s.SleepMax
	if r.SleepMax > max {
		max = r.SleepMax
	}
	return &Signature{
		State:     s.State,     // Drop right side.
		CreatedBy: s.CreatedBy, // Drop right side.
		SleepMin:  min,
		SleepMax:  max,
		Stack:     *s.Stack.Merge(&r.Stack),
		Locked:    s.Locked || r.Locked, // TODO(maruel): This is weirdo.
	}
}

// Less compares two Signature, where the ones that are less are more
// important, so they come up front. A Signature with more private functions is
// 'less' so it is at the top. Inversely, a Signature with only public
// functions is 'more' so it is at the bottom.
func (s *Signature) Less(r *Signature) bool {
	if s.Stack.Less(&r.Stack) {
		return true
	}
	if r.Stack.Less(&s.Stack) {
		return false
	}
	if s.Locked && !r.Locked {
		return true
	}
	if r.Locked && !s.Locked {
		return false
	}
	if s.State < r.State {
		return true
	}
	if s.State > r.State {
		return false
	}
	return false
}

// Goroutine represents the state of one goroutine, including the stack trace.
type Goroutine struct {
	Signature      // It's stack trace, internal bits, state, which call site created it, etc.
	ID        int  // Goroutine ID.
	First     bool // First is the goroutine first printed, normally the one that crashed.
}

// Bucketize returns the number of similar goroutines.
func Bucketize(goroutines []Goroutine, similar Similarity) map[*Signature][]Goroutine {
	out := map[*Signature][]Goroutine{}
	// O(n²). Fix eventually.
	for _, routine := range goroutines {
		found := false
		for key := range out {
			// When a match is found, this effectively drops the other goroutine ID.
			if key.Similar(&routine.Signature, similar) {
				found = true
				if !key.Equal(&routine.Signature) {
					// Almost but not quite equal. There's different pointers passed
					// around but the same values. Zap out the different values.
					newKey := key.Merge(&routine.Signature)
					out[newKey] = append(out[key], routine)
					delete(out, key)
				} else {
					out[key] = append(out[key], routine)
				}
				break
			}
		}
		if !found {
			key := &Signature{}
			*key = routine.Signature
			out[key] = []Goroutine{routine}
		}
	}
	return out
}

// Bucket is a stack trace signature and the list of goroutines that fits this
// signature.
type Bucket struct {
	Signature
	Routines []Goroutine
}

// First returns true if it contains the first goroutine, e.g. the ones that
// likely generated the panic() call, if any.
func (b *Bucket) First() bool {
	for _, r := range b.Routines {
		if r.First {
			return true
		}
	}
	return false
}

// Less does reverse sort.
func (b *Bucket) Less(r *Bucket) bool {
	if b.First() {
		return true
	}
	if r.First() {
		return false
	}
	return b.Signature.Less(&r.Signature)
}

// Buckets is a list of Bucket sorted by repeation count.
type Buckets []Bucket

func (b Buckets) Len() int {
	return len(b)
}

func (b Buckets) Less(i, j int) bool {
	return b[i].Less(&b[j])
}

func (b Buckets) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

// SortBuckets creates a list of Bucket from each goroutine stack trace count.
func SortBuckets(buckets map[*Signature][]Goroutine) Buckets {
	out := make(Buckets, 0, len(buckets))
	for signature, count := range buckets {
		out = append(out, Bucket{*signature, count})
	}
	sort.Sort(out)
	return out
}

// scanLines is similar to bufio.ScanLines except that it:
//     - doesn't drop '\n'
//     - doesn't strip '\r'
//     - returns when the data is bufio.MaxScanTokenSize bytes
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i + 1, data[0 : i+1], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	if len(data) >= bufio.MaxScanTokenSize {
		// Returns the line even if it is not at EOF nor has a '\n', otherwise the
		// scanner will return bufio.ErrTooLong which is definitely not what we
		// want.
		return len(data), data, nil
	}
	return 0, nil, nil
}

// ParseDump processes the output from runtime.Stack().
//
// It supports piping from another command and assumes there is junk before the
// actual stack trace. The junk is streamed to out.
func ParseDump(r io.Reader, out io.Writer) ([]Goroutine, error) {
	goroutines := make([]Goroutine, 0, 16)
	var goroutine *Goroutine
	scanner := bufio.NewScanner(r)
	scanner.Split(scanLines)
	// TODO(maruel): Use a formal state machine. Patterns follows:
	// - reRoutineHeader
	//   Either:
	//     - reUnavail
	//     - reFunc + reFile in a loop
	//     - reElided
	//   Optionally ends with:
	//     - reCreated + reFile
	// Between each goroutine stack dump: an empty line
	created := false
	// firstLine is the first line after the reRoutineHeader header line.
	firstLine := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == "\n" {
			if goroutine != nil {
				goroutine = nil
				continue
			}
		} else if line[len(line)-1] == '\n' {
			if goroutine == nil {
				if match := reRoutineHeader.FindStringSubmatch(line); match != nil {
					if id, err := strconv.Atoi(match[1]); err == nil {
						// See runtime/traceback.go.
						// "<state>, \d+ minutes, locked to thread"
						items := strings.Split(match[2], ", ")
						sleep := 0
						locked := false
						for i := 1; i < len(items); i++ {
							if items[i] == lockedToThread {
								locked = true
								continue
							}
							// Look for duration, if any.
							if match2 := reMinutes.FindStringSubmatch(items[i]); match2 != nil {
								sleep, _ = strconv.Atoi(match2[1])
							}
						}
						goroutines = append(goroutines, Goroutine{
							Signature: Signature{
								State:    items[0],
								SleepMin: sleep,
								SleepMax: sleep,
								Locked:   locked,
							},
							ID:    id,
							First: len(goroutines) == 0,
						})
						goroutine = &goroutines[len(goroutines)-1]
						firstLine = true
						continue
					}
				}
			} else {
				if firstLine {
					firstLine = false
					if match := reUnavail.FindStringSubmatch(line); match != nil {
						// Generate a fake stack entry.
						goroutine.Stack.Calls = []Call{{SourcePath: "<unavailable>"}}
						continue
					}
				}

				if match := reFile.FindStringSubmatch(line); match != nil {
					// Triggers after a reFunc or a reCreated.
					num, err := strconv.Atoi(match[2])
					if err != nil {
						return goroutines, fmt.Errorf("failed to parse int on line: \"%s\"", line)
					}
					if created {
						created = false
						goroutine.CreatedBy.SourcePath = match[1]
						goroutine.CreatedBy.Line = num
					} else {
						i := len(goroutine.Stack.Calls) - 1
						if i < 0 {
							return goroutines, errors.New("unexpected order")
						}
						goroutine.Stack.Calls[i].SourcePath = match[1]
						goroutine.Stack.Calls[i].Line = num
					}
					continue
				}

				if match := reCreated.FindStringSubmatch(line); match != nil {
					created = true
					goroutine.CreatedBy.Func.Raw = match[1]
					continue
				}

				if match := reFunc.FindStringSubmatch(line); match != nil {
					args := Args{}
					for _, a := range strings.Split(match[2], ", ") {
						if a == "..." {
							args.Elided = true
							continue
						}
						if a == "" {
							// Remaining values were dropped.
							break
						}
						v, err := strconv.ParseUint(a, 0, 64)
						if err != nil {
							return goroutines, fmt.Errorf("failed to parse int on line: \"%s\"", line)
						}
						args.Values = append(args.Values, Arg{Value: v})
					}
					goroutine.Stack.Calls = append(goroutine.Stack.Calls, Call{Func: Function{match[1]}, Args: args})
					continue
				}

				if match := reElided.FindStringSubmatch(line); match != nil {
					goroutine.Stack.Elided = true
					continue
				}
			}
		}
		_, _ = io.WriteString(out, line)
		goroutine = nil
	}
	nameArguments(goroutines)
	return goroutines, scanner.Err()
}

// Private stuff.

func nameArguments(goroutines []Goroutine) {
	// Set a name for any pointer occuring more than once.
	type object struct {
		args      []*Arg
		inPrimary bool
		id        int
	}
	objects := map[uint64]object{}
	// Enumerate all the arguments.
	for i := range goroutines {
		for j := range goroutines[i].Stack.Calls {
			for k := range goroutines[i].Stack.Calls[j].Args.Values {
				arg := goroutines[i].Stack.Calls[j].Args.Values[k]
				if arg.IsPtr() {
					objects[arg.Value] = object{
						args:      append(objects[arg.Value].args, &goroutines[i].Stack.Calls[j].Args.Values[k]),
						inPrimary: objects[arg.Value].inPrimary || i == 0,
					}
				}
			}
		}
		// CreatedBy.Args is never set.
	}
	order := uint64Slice{}
	for k, obj := range objects {
		if len(obj.args) > 1 && obj.inPrimary {
			order = append(order, k)
		}
	}
	sort.Sort(order)
	nextID := 1
	for _, k := range order {
		for _, arg := range objects[k].args {
			arg.Name = fmt.Sprintf("#%d", nextID)
		}
		nextID++
	}

	// Now do the rest. This is done so the output is deterministic.
	order = uint64Slice{}
	for k := range objects {
		order = append(order, k)
	}
	sort.Sort(order)
	for _, k := range order {
		// Process the remaining pointers, they were not referenced by primary
		// thread so will have higher IDs.
		if objects[k].inPrimary {
			continue
		}
		for _, arg := range objects[k].args {
			arg.Name = fmt.Sprintf("#%d", nextID)
		}
		nextID++
	}
}

type uint64Slice []uint64

func (a uint64Slice) Len() int           { return len(a) }
func (a uint64Slice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a uint64Slice) Less(i, j int) bool { return a[i] < a[j] }
