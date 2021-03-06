/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package server

import (
	dubboclient "github.com/apache/servicecomb-mesher/proxy/protocol/dubbo/client"
	"github.com/apache/servicecomb-mesher/proxy/protocol/dubbo/dubbo"
	"sync"
	"time"

	//_ "github.com/apache/servicecomb-mesher/proxy/protocol/dubbo/client"
	"github.com/go-chassis/go-chassis/v2/core/config"
	"github.com/go-chassis/go-chassis/v2/core/config/model"
	"github.com/go-chassis/go-chassis/v2/core/lager"
	"github.com/go-chassis/go-chassis/v2/core/server"
	//"github.com/go-chassis/go-chassis/v2/core/client
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	lager.Init(&lager.Options{LoggerLevel: "DEBUG"})
}

func TestDubboServer_Start(t *testing.T) {
	//config.Init()

	protoMap := make(map[string]model.Protocol)
	config.GlobalDefinition = &model.GlobalCfg{
		ServiceComb: model.ServiceComb{
			Protocols: protoMap,
		},
	}

	defaultChain := make(map[string]string)
	defaultChain["default"] = ""

	config.GlobalDefinition.ServiceComb.Handler.Chain.Provider = defaultChain
	config.GlobalDefinition.ServiceComb.Handler.Chain.Consumer = defaultChain
	config.MicroserviceDefinition = &model.ServiceSpec{}

	f, err := server.GetServerFunc("dubbo")
	assert.NoError(t, err)

	// case split port error
	s := f(server.Options{
		Address:   "0.0.0.130201",
		ChainName: "default",
	})
	err = s.Start()
	assert.Error(t, err)

	// case invalid host
	s = f(server.Options{
		Address:   "2.2.2.1990:30201",
		ChainName: "default",
	})
	err = s.Start()
	assert.Error(t, err)

	// case listening error
	s = f(server.Options{
		Address:   "99.0.0.1:30201",
		ChainName: "default",
	})
	err = s.Start()
	assert.Error(t, err)

	// case ok
	s = f(server.Options{
		Address:   "127.0.0.1:30201",
		ChainName: "default",
	})
	err = s.Start()
	assert.NoError(t, err)

	s.Stop()
	time.Sleep(time.Second * 5)
}

func TestDubboServer(t *testing.T) {
	t.Log("Test dubbo server function")
	config.Init()

	protoMap := make(map[string]model.Protocol)
	config.GlobalDefinition = &model.GlobalCfg{
		ServiceComb: model.ServiceComb{
			Protocols: protoMap,
		},
	}

	defaultChain := make(map[string]string)
	defaultChain["default"] = ""

	config.GlobalDefinition.ServiceComb.Handler.Chain.Provider = defaultChain
	config.GlobalDefinition.ServiceComb.Handler.Chain.Consumer = defaultChain
	config.MicroserviceDefinition = &model.ServiceSpec{}

	f, err := server.GetServerFunc("dubbo")
	assert.NoError(t, err)
	addr := "127.0.0.1:40201"
	s := f(server.Options{
		Address:   addr,
		ChainName: "default",
	})

	s.Register(map[string]string{})

	err = s.Start()
	assert.NoError(t, err)

	name := s.String()
	assert.Equal(t, "dubbo", name)

	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		clientMgr := dubboclient.NewClientMgr()
		var dubboClient *dubboclient.DubboClient
		dubboClient, err := clientMgr.GetClient(addr, time.Second*5)
		assert.NoError(t, err)

		req := new(dubbo.Request)
		req.SetMsgID(int64(11111111))
		req.SetVersion("1.0.0")
		req.SetEvent("ok")

		_, err = dubboClient.Send(req)

	}(&wg)

	wg.Wait()

	err = s.Stop()
	assert.NoError(t, err)
}
