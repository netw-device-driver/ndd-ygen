/*
Copyright 2020 Wim Henderickx.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package container

type Container struct {
	Name    string
	Entries []*Entry
	Prev    *Container
}

// Entry structure keeps track of the elements in a struct/list
type Entry struct {
	Next          *Container
	Prev          *Container
	Name          string
	Type          string
	Enum          []string
	EnumString    string
	Range         []int
	Length        []int
	Pattern       []string
	PatternString string
	Union         bool
	Mandatory     bool
	Default       string
	Key           string
}

type ContainerOption func(c *Container)

func NewContainer(n string, prev *Container, opts ...ContainerOption) *Container {
	e := &Container{
		Name:    n,
		Entries: make([]*Entry, 0),
		Prev:    prev,
	}

	for _, o := range opts {
		o(e)
	}

	return e
}

func (c *Container) GetFullName() string {
	if c.Prev != nil {
		return getRecursiveName(c.Prev) + "-" + c.Name
	}
	return c.Name
}

func getRecursiveName(c *Container) string {
	if c.Prev != nil {
		return getRecursiveName(c.Prev) + "-" + c.Name
	}
	return c.Name
}

// Option can be used to manipulate Options.
type EntryOption func(c *Entry)

func WithType(s string) EntryOption {
	return func(c *Entry) {
		c.Type = s
	}
}

func WithEnum(s []string) EntryOption {
	return func(c *Entry) {
		c.Enum = s
	}
}

func WithRange(s []int) EntryOption {
	return func(c *Entry) {
		c.Range = s
	}
}

func WithLength(s []int) EntryOption {
	return func(c *Entry) {
		c.Length = s
	}
}

func WithPattern(s []string) EntryOption {
	return func(c *Entry) {
		c.Pattern = s
	}
}

func WithUnion(b bool) EntryOption {
	return func(c *Entry) {
		c.Union = b
	}
}

func WithMandatory(b bool) EntryOption {
	return func(c *Entry) {
		c.Mandatory = b
	}
}

func WithDefault(s string) EntryOption {
	return func(c *Entry) {
		c.Default = s
	}
}

/*
func WithKey(s string) Option {
	return func(c *Entry) {
		c.Key = s
	}
}

func WithKeyType(s string) Option {
	return func(c *Entry) {
		c.KeyType = s
	}
}
*/

func NewEntry(n string, opts ...EntryOption) *Entry {
	if n == "ethernet-segment" {
		n = "esi"
	}
	e := &Entry{
		Name:    n,
		Next:    nil,
		Prev:    nil,
		Enum:    make([]string, 0),
		Range:   make([]int, 0),
		Length:  make([]int, 0),
		Pattern: make([]string, 0),
	}

	for _, o := range opts {
		o(e)
	}

	return e
}