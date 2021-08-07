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

package nddygen

import (
	"github.com/netw-device-driver/ndd-ygen/pkg/generator"
	"github.com/netw-device-driver/ndd-runtime/pkg/logging"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var yangImportDirs []string
var yangModuleDirs []string
var resourceMapInputFile string
var outputDir string
var packageName string
var version string
var prefix string
var apiGroup string

const (
	errCreateGenerator = "cannot initialize generator"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:          "generate",
	Short:        "generate ndd provider using yang",
	Aliases:      []string{"gen"},
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		zlog := zap.New(zap.UseDevMode(debug), zap.JSONEncoder())
		log := logging.NewLogrLogger(zlog.WithName("nddgenyang"))
		log.Debug("generate provider ...")

		opts := []generator.Option{
			generator.WithYangImportDirs(yangImportDirs),
			generator.WithYangModuleDirs(yangModuleDirs),
			generator.WithResourceMapInputFile(resourceMapInputFile),
			generator.WithOutputDir(outputDir),
			generator.WithPackageName(packageName),
			generator.WithVersion(version),
			generator.WithAPIGroup(apiGroup),
			generator.WithPrefix(prefix),
			generator.WithLogging(log),
			generator.WithDebug(debug),
		}
		g, err := generator.NewGenerator(opts...)
		if err != nil {
			return errors.Wrap(err, errCreateGenerator)
		}
		//g.ShowConfiguration()

		if err := g.Run(); err != nil {
			log.Debug("Error", "error", err)
			return err
		}
		if err := g.Render(); err != nil {
			log.Debug("Error", "error", err)
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringSliceVarP(&yangImportDirs, "yang-import-dirs", "i", []string{"/Users/henderiw/CodeProjects/go-dev/ndd-ygen/conf/yang/21_03_0/ietf/"}, "Comma separated list of dirs to be recursively searched for import modules.")
	generateCmd.Flags().StringSliceVarP(&yangModuleDirs, "yang-module-dirs", "m", []string{"/Users/henderiw/CodeProjects/go-dev/ndd-ygen/conf/yang/21_03_0/srl/"}, "Comma separated list of dirs to be recursively searched for yang modules")
	generateCmd.Flags().StringVarP(&resourceMapInputFile, "resource-map-input", "r", "/Users/henderiw/CodeProjects/go-dev/ndd-ygen/conf/resourceMapInputPlayK8s.yaml", "The resource map input file which resource should be generated")
	generateCmd.Flags().StringVarP(&outputDir, "output-dir", "o", "out/", "The directory that the Go package should be written to.")
	generateCmd.Flags().StringVarP(&packageName, "package-name", "p", "tfsrl", "The packageName the code will generate")
	generateCmd.Flags().StringVarP(&version, "version", "v", "v1", "The version of the api to geenrate")
	generateCmd.Flags().StringVarP(&apiGroup, "apiGroup", "g", "srl.ndd.henderiw.be", "The group of the api to geenrate")
	generateCmd.Flags().StringVarP(&prefix, "prefix", "a", "srl", "The prefix that is added to the kubernetes api resource")
	
}
