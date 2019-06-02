// Copyright © 2019.  Douglas Chimento <dchimento@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"

	awsevents "github.com/aws/aws-lambda-go/events"
	"github.com/dougEfresh/passwd-pot/resolver"
)

type IpRequest struct {
	IpAddr string `json:ipAddr`

}

func HandleGeoLookup(ctx context.Context, request IpRequest) (resolver.Geo, error) {
	if request.IpAddr == "" {
		return resolver.Geo{}, APIError{GatewayError: awsevents.APIGatewayProxyResponse{StatusCode: 500, Headers: header, Body: "Invalid IpAddr"}}
	}

	geo, err := geoClient.GetLocationForAddr(request.IpAddr)
	if err != nil {
		return resolver.Geo{}, APIError{GatewayError: awsevents.APIGatewayProxyResponse{StatusCode: 500, Headers: header, Body: err.Error()}}
	}
	return *geo, nil
}
