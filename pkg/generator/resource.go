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

package generator

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	config "github.com/netw-device-driver/ndd-grpc/config/configpb"
	"github.com/netw-device-driver/ndd-runtime/pkg/yang/container"
	"github.com/netw-device-driver/ndd-runtime/pkg/yang/parser"
	"github.com/netw-device-driver/ndd-runtime/pkg/yang/resource"
	"github.com/openconfig/goyang/pkg/yang"
)

// FindBestMatch finds the string which matches the most
func (g *Generator) FindBestMatch(path config.Path) (*resource.Resource, bool) {
	minLength := 0
	resMatch := &resource.Resource{}
	found := false
	for _, r := range g.Resources {
		if strings.Contains(*parser.GnmiPathToXPath(&path, false), *r.GetAbsoluteXPath()) {
			// find the string which matches the most
			// should be the last match normally since we added them
			// to the list from root to lower hierarchy

			if len([]rune(*r.GetAbsoluteXPath())) > minLength {
				minLength = len([]rune(*r.GetAbsoluteXPath()))
				resMatch = r
				found = true
			}
		}
	}
	return resMatch, found
}

// IsResourcesInit checks if the resource exists
func (g *Generator) DoesResourceMatch(path config.Path) (*resource.Resource, bool) {
	//fmt.Printf("Path: %s\n", *parser.GnmiPathToXPath(path))
	if r, ok := g.FindBestMatch(path); ok {
		//fmt.Printf("match path: %s \n", *r.GetAbsoluteXPath())
		return r, true

	}
	return nil, false
}

func (g *Generator) ResourceGenerator(resPath string, path config.Path, e *yang.Entry) error {
	resPath += filepath.Join("/", e.Name)
	//fmt.Printf("resource path1: %s \n", resPath)
	path.Elem = append(path.Elem, parser.InitializePathElem(e))
	//fmt.Printf("resource path2: %s \n", *parser.GnmiPathToXPath(&path, false))

	if r, ok := g.DoesResourceMatch(path); ok {
		//fmt.Printf("match path: %s \n", *r.GetAbsoluteXPath())
		switch {
		case e.RPC != nil:
		case e.ReadOnly():
		default: // this is a RW config element in yang
			// find the containerPointer
			// we look at the level delta from the root of the resource -> newLevel
			// newLevel = 0 is special since it is the root of the container
			// newLevel = 0 since there is no container yet we cannot find the container Pointer, since it is not created so far
			newLevel := strings.Count(resPath, "/") - strings.Count(*r.GetAbsoluteXPathWithoutKey(), "/")
			var cPtr *container.Container
			if newLevel > 0 {
				r.ContainerLevel = newLevel

				cPtr = r.ContainerLevelKeys[newLevel-1][len(r.ContainerLevelKeys[newLevel-1])-1]
			}
			fmt.Printf("xpath: %s, resPath: %s, level: %d\n", *r.GetAbsoluteXPathWithoutKey(), resPath, r.ContainerLevel)

			// Leaf processing
			if e.Kind.String() == "Leaf" {
				fmt.Printf("Leaf Name: %s, ResPath: %s \n", e.Name, resPath)
				// add entry to the container
				cPtr.Entries = append(cPtr.Entries, parser.CreateContainerEntry(e, nil, nil))
				localPath, remotePath, local := parser.ProcessLeafRef(e, resPath, r.GetAbsoluteGnmiActualResourcePath())
				if localPath != nil {
					// validate if the leafrefs is a local leafref or an externaal leafref
					if local {
						// local leafref
						r.AddLocalLeafRef(localPath, remotePath)
					} else {
						// external leafref
						r.AddExternalLeafRef(localPath, remotePath)
					}
				}
			} else { // List processing with or without a key
				fmt.Printf("List Name: %s, ResPath: %s \n", e.Name, resPath)
				// newLevel = 0 is special since we have to initialize the container
				// for newLevl = 0 we do not have to rely on the cPtr, since there is no cPtr initialized yet
				// for newLevl = 0 we dont create an entry in the container but we create a root container entry
				if newLevel == 0 {
					// create a new container and apply to the root of the resource
					r.Container = container.NewContainer(e.Name, nil)
					// r.Container.Entries = append(r.Container.Entries, parser.CreateContainerEntry(e, nil, nil))
					// append the container Ptr to the back of the list, to track the used container Pointers per level
					// newLevel =0
					r.SetRootContainerEntry(parser.CreateContainerEntry(e, nil, nil))
					r.ContainerLevelKeys[newLevel] = make([]*container.Container, 0)
					r.ContainerLevelKeys[newLevel] = append(r.ContainerLevelKeys[newLevel], r.Container)
					r.ContainerList = append(r.ContainerList, r.Container)

				} else {
					// create a new container for the next iteration
					c := container.NewContainer(e.Name, cPtr)
					// allocate container entry to the original container Pointer and append to the container entry list
					// the next pointer of the entry points to the new container
					cPtr.Entries = append(cPtr.Entries, parser.CreateContainerEntry(e, c, cPtr))
					// append the container Ptr to the back of the list, to track the used container Pointers per level
					// initialize the level
					r.ContainerLevelKeys[newLevel] = make([]*container.Container, 0)
					r.ContainerLevelKeys[newLevel] = append(r.ContainerLevelKeys[newLevel], c)
					r.ContainerList = append(r.ContainerList, c)

				}

			}
		}
	}
	// handles the recursive analysis of the yang tree
	var names []string
	for k := range e.Dir {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		g.ResourceGenerator(resPath, path, e.Dir[k])
	}
	return nil
}
