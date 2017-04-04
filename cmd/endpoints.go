// Copyright © 2017 Douglas Chimento <dchimento@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	RecordEndpoint endpoint.Endpoint
}

func MakeServerEndpoints(s EventService) Endpoints {
	return Endpoints{
		MakeRecordEndpoint(s),
	}
}

func MakeRecordEndpoint(s EventService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(Event)
		return s.Record(ctx, req)
	}
}
