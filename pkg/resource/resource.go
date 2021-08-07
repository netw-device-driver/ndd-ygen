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

package resource

import (
	"os"
	"path/filepath"

	"github.com/netw-device-driver/ndd-ygen/pkg/container"
	"github.com/netw-device-driver/ndd-ygen/pkg/leafref"
	"github.com/netw-device-driver/ndd-ygen/pkg/parser"
	config "github.com/netw-device-driver/ndd-grpc/config/configpb"
	"github.com/stoewer/go-strcase"
)

type Resource struct {
	Path               *config.Path                   // relative path from the resource; the absolute path is assembled using the resurce hierarchy with dependsOn
	DependsOn          *Resource                      // resource dependency
	Excludes           []*config.Path                 // relative from the the resource
	FileName           string                         // the filename the resource is using to render out the config
	ResFile            *os.File                       // the file reference for writing the resource file
	RootContainerEntry *container.Entry               // this is the root element which is used to reference the hierarchical resource information
	Container          *container.Container           // root container of the resource
	LastContainerPtr   *container.Container           // pointer to the last container we process
	ContainerList      []*container.Container         // List of all containers within the resource
	ContainerLevel     int                            // the current container Level when processing the yang entries
	ContainerLevelKeys map[int][]*container.Container // the current container Level key list
	LocalLeafRefs      []*leafref.LeafRef
	ExternalLeafRefs   []*leafref.LeafRef
}

// Option can be used to manipulate Options.
type Option func(g *Resource)

func WithXPath(p string) Option {
	return func(r *Resource) {
		r.Path = parser.XpathToGnmiPath(p, 0)
	}
}

func WithDependsOn(d *Resource) Option {
	return func(r *Resource) {
		r.DependsOn = d
	}
}

func WithExclude(p string) Option {
	return func(r *Resource) {
		r.Excludes = append(r.Excludes, parser.XpathToGnmiPath(p, 0))
	}
}

func NewResource(opts ...Option) *Resource {
	r := &Resource{
		Path: new(config.Path),
		//DependsOn:          new(Resource),
		Excludes:           make([]*config.Path, 0),
		RootContainerEntry: nil,
		Container:          nil,
		LastContainerPtr:   nil,
		ContainerList:      make([]*container.Container, 0),
		ContainerLevel:     0,
		ContainerLevelKeys: make(map[int][]*container.Container),
		LocalLeafRefs:      make([]*leafref.LeafRef, 0),
		ExternalLeafRefs:   make([]*leafref.LeafRef, 0),
	}

	for _, o := range opts {
		o(r)
	}

	r.ContainerLevelKeys[0] = make([]*container.Container, 0)

	return r
}

func (r *Resource) AddLocalLeafRef(ll, rl *config.Path) {
	r.LocalLeafRefs = append(r.LocalLeafRefs, &leafref.LeafRef{
		LocalPath:  ll,
		RemotePath: rl,
	})
}

func (r *Resource) AddExternalLeafRef(ll, rl *config.Path) {
	r.ExternalLeafRefs = append(r.ExternalLeafRefs, &leafref.LeafRef{
		LocalPath:  ll,
		RemotePath: rl,
	})
}

func (r *Resource) GetResourceNameWithPrefix(prefix string) string {
	return strcase.UpperCamelCase(prefix + "-" + r.GetAbsoluteName())
}

func (r *Resource) AssignFileName(prefix, suffix string) {
	r.FileName = prefix + "-" + strcase.KebabCase(r.GetAbsoluteName()) + suffix
}

func (r *Resource) CreateFile(dir, subdir1, subdir2 string) (err error) {
	r.ResFile, err = os.Create(filepath.Join(dir, subdir1, subdir2, filepath.Base(r.FileName)))
	return err
}

func (r *Resource) CloseFile() error {
	return r.ResFile.Close()
}

func (r *Resource) ResourceLastElement() string {
	return r.Path.GetElem()[len(r.Path.GetElem())-1].GetName()
}

func (r *Resource) GetRelativeGnmiPath() *config.Path {
	return r.Path
}

// root resource have a additional entry in the path which is inconsistent with hierarchical resources
// to provide consistencyw e introduced this method to provide a consistent result for paths
// used mainly for leafrefs for now
func (r *Resource) GetRelativeGnmiActualResourcePath() *config.Path {
	if r.DependsOn != nil {
		return r.Path
	}
	actPath := *r.Path
	actPath.Elem = actPath.Elem[1:(len(actPath.GetElem()))]
	return &actPath
}

func (r *Resource) GetRelativeXPath() *string {
	return parser.GnmiPathToXPath(r.Path, true)
}

func (r *Resource) GetAbsoluteName() string {
	e := findPathElemHierarchy(r)
	if len(e) > 1 {
		e = e[1:]
	}
	return parser.GnmiPathToName(&config.Path{
		Elem: e,
	})
}

// root resource have a additional entry in the path which is inconsistent with hierarchical resources
// to provide consistency we introduced this method to provide a consistent result for paths
// used mainly for leafrefs for now
func (r *Resource) GetAbsoluteGnmiActualResourcePath() *config.Path {
	actPath := &config.Path{
		Elem: findPathElemHierarchy(r),
	}

	actPath.Elem = actPath.Elem[1:(len(actPath.GetElem()))]
	return actPath
}

func (r *Resource) GetAbsoluteGnmiPath() *config.Path {
	return &config.Path{
		Elem: findPathElemHierarchy(r),
	}
}

func (r *Resource) GetAbsoluteXPathWithoutKey() *string {
	return parser.GnmiPathToXPath(&config.Path{
		Elem: findPathElemHierarchy(r),
	}, false)
}

func (r *Resource) GetAbsoluteXPath() *string {
	return parser.GnmiPathToXPath(&config.Path{
		Elem: findPathElemHierarchy(r),
	}, true)
}

func (r *Resource) GetExcludeRelativeXPath() []string {
	e := make([]string, 0)
	for _, p := range r.Excludes {
		e = append(e, *parser.GnmiPathToXPath(p, true))
	}
	return e
}

func findPathElemHierarchy(r *Resource) []*config.PathElem {
	if r.DependsOn != nil {
		fp := findPathElemHierarchy(r.DependsOn)
		fp = append(fp, r.Path.Elem...)
		return fp
	}
	return r.Path.GetElem()
}

func (r *Resource) SetRootContainerEntry(e *container.Entry) {
	r.RootContainerEntry = e
}

func (r *Resource) GetHierarchicalElements() []*HeInfo {
	he := make([]*HeInfo, 0)
	if r.DependsOn != nil {
		he = findHierarchicalElements(r.DependsOn, he)
	}
	return he
}

func findHierarchicalElements(r *Resource, he []*HeInfo) []*HeInfo {
	h := &HeInfo{
		Name: r.RootContainerEntry.Name,
		Key: r.RootContainerEntry.Key,
		Type: r.RootContainerEntry.Type,
	}
	he = append(he, h)
	if r.DependsOn != nil {
		he = findHierarchicalElements(r.DependsOn, he)
	}
	return he
}

type HeInfo struct {
	Name string
	Key  string
	Type string
}
