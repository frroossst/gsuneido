package globals

import (
	"sync"

	"github.com/apmckinlay/gsuneido/util/verify"
	"github.com/apmckinlay/gsuneido/value"
)

var (
	lock     sync.RWMutex
	name2num = make(map[string]int)
	// put nil in first slot so we never use gnum of zero
	names  = []string{""}
	values = []value.Value{nil}
)

// Add adds a new name and value to globals.
//
// This is used for set up of built-in globals
// The return value is so it can be used like:
// var _ = globals.Add(...)
func Add(name string, val value.Value) int {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := name2num[name]; ok {
		panic("duplicate global: " + name)
	}
	gnum := len(names)
	name2num[name] = gnum
	names = append(names, name)
	values = append(values, val)
	verify.That(len(names) == len(values))
	return gnum
}

// NameNum returns the global number for a name
// adding it if it doesn't exist.
func NameNum(name string) int {
	gn, ok := check(name)
	if ok {
		return gn
	}
	return Add(name, nil)
}

func check(name string) (int, bool) {
	lock.RLock()
	defer lock.RUnlock()
	gn, ok := name2num[name]
	return gn, ok
}

// NumName returns the name for a global number
func NumName(gnum int) string {
	lock.RLock()
	defer lock.RUnlock()
	return names[gnum]
}

// Get returns the value for a global
func Get(gnum int) value.Value {
	lock.RLock()
	defer lock.RUnlock()
	return values[gnum]
}
