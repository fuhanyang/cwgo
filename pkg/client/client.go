/*
 * Copyright 2022 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"bytes"
	"os"
	"strings"

	"github.com/cloudwego/cwgo/pkg/common/kx_registry"
	"github.com/cloudwego/cwgo/pkg/consts"

	"github.com/cloudwego/cwgo/pkg/common/utils"
	kargs "github.com/cloudwego/kitex/tool/cmd/kitex/args"
	"github.com/cloudwego/kitex/tool/internal_pkg/log"
	"github.com/cloudwego/kitex/tool/internal_pkg/pluginmode/thriftgo"

	"github.com/cloudwego/hertz/cmd/hz/app"

	"github.com/cloudwego/cwgo/config"

	hzConfig "github.com/cloudwego/hertz/cmd/hz/config"
	"github.com/cloudwego/hertz/cmd/hz/meta"
	"github.com/cloudwego/hertz/cmd/hz/util/logs"
	"github.com/urfave/cli/v2"
)

func Client(c *config.ClientArgument) error {
	var err error
	err = check(c)
	if err != nil {
		return err
	}
	switch c.Type {
	case consts.RPC:
		log.Verbose = c.Verbose

		idls, err := utils.ExpandIDLPaths(c.IdlPath)
		if err != nil {
			return err
		}

		origServerName := c.ServerName
		for i, idl := range idls {
			cc := *c
			common := *c.CommonParam
			cc.CommonParam = &common
			cc.IdlPath = idl
			slice := *c.SliceParam
			slice.Pass = append([]string(nil), c.SliceParam.Pass...)
			slice.ProtoSearchPath = append([]string(nil), c.SliceParam.ProtoSearchPath...)
			cc.SliceParam = &slice

			if i > 0 {
				cc.ServerName = ""
			} else {
				cc.ServerName = origServerName
			}

			var args kargs.Arguments
			err = convertKitexArgs(&cc, &args)
			if err != nil {
				return err
			}

			kx_registry.HandleRegistry(cc.CommonParam, args.TemplateDir)
			defer kx_registry.RemoveExtension()

			out := new(bytes.Buffer)
			cmd := args.BuildCmd(out)
			err = cmd.Run()
			if err != nil {
				if args.Use != "" {
					out := strings.TrimSpace(out.String())
					if strings.HasSuffix(out, thriftgo.TheUseOptionMessage) {
						utils.ReplaceThriftVersion()
					}
				}
				os.Exit(1)
			}
			utils.Hessian2PostProcessing(args)
		}

		utils.ReplaceThriftVersion()
		utils.UpgradeGolangProtobuf()
	case consts.HTTP:
		args := hzConfig.NewArgument()
		utils.SetHzVerboseLog(c.Verbose)
		err = convertHzArgument(c, args)
		if err != nil {
			return err
		}
		args.CmdType = meta.CmdClient
		logs.Debugf("Args: %#v\n", args)
		err = app.TriggerPlugin(args)
		if err != nil {
			return cli.Exit(err, meta.PluginError)
		}
	}
	return nil
}
