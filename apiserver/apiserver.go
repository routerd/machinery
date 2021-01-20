/*
routerd
Copyright (C) 2020  The routerd Authors

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package apiserver

import (
	"context"
	"crypto/tls"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type APIServer struct {
	s           *http.Server
	config      APIServerConfig
	registerFns []RegisterFn

	grpcServer     *grpc.Server
	grpcGatewayMux *gwruntime.ServeMux
	grpcClient     *grpc.ClientConn
}

type APIServerConfig struct {
	Address                 string
	TLSCertFile, TLSKeyFile string
}

type RegisterFn func(
	ctx context.Context, grpcServer *grpc.Server,
	grpcGatewayMux *gwruntime.ServeMux, grpcClient *grpc.ClientConn) error

func NewServer(c APIServerConfig, registerFns ...RegisterFn) (*APIServer, error) {
	grpcServer := grpc.NewServer()
	grpcWebServer := grpcweb.WrapServer(grpcServer)
	grpcGatewayMux := gwruntime.NewServeMux(
	// gwruntime.WithProtoErrorHandler
	// gwruntime.
	// gwruntime.WithProtoErrorHandler(func(
	// 	ctx context.Context,
	// 	serveMux *gwruntime.ServeMux,
	// 	marshaler gwruntime.Marshaler,
	// 	writer http.ResponseWriter,
	// 	request *http.Request, err error) {
	// 	const fallback = `{"error": "failed to marshal error message"}`
	// 	writer.Header().Del("Trailer")
	// 	writer.Header().Set("Content-Type", marshaler.ContentType())
	// 	s, ok := status.FromError(err)
	// 	if !ok {
	// 		s = status.New(codes.Unknown, err.Error())
	// 	}
	// 	buf, marshalerr := marshaler.Marshal(s.Proto())
	// 	if marshalerr != nil {
	// 		writer.WriteHeader(http.StatusInternalServerError)
	// 		_, _ = io.WriteString(writer, fallback)
	// 		return
	// 	}

	// 	st := gwruntime.HTTPStatusFromCode(s.Code())
	// 	writer.WriteHeader(st)
	// 	_, _ = writer.Write(buf)
	// }),
	)
	// gwruntime.SetHTTPBodyMarshaler(grpcGatewayMux)

	var handler http.Handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.Contains(request.Header.Get("Content-Type"), "application/grpc") {
			grpcWebServer.ServeHTTP(writer, request)
		} else {
			grpcGatewayMux.ServeHTTP(writer, request)
		}
	})
	handler = handlers.CORS(
		handlers.AllowedHeaders([]string{
			"X-Requested-With",
			"Content-Type",
			"Authorization",
			"X-grpc-web",
			"X-user-agent",
		}),
		handlers.AllowedMethods([]string{"GET", "POST"}),
		handlers.AllowedOrigins([]string{"*"}),
	)(handler)

	grpcClient, err := grpc.Dial(c.Address,
		grpc.WithTransportCredentials(
			credentials.NewTLS(&tls.Config{
				InsecureSkipVerify: true,
			})))
	if err != nil {
		return nil, err
	}

	apiserver := &APIServer{
		s: &http.Server{
			Handler: handler,
			Addr:    c.Address,
		},
		config:         c,
		registerFns:    registerFns,
		grpcServer:     grpcServer,
		grpcClient:     grpcClient,
		grpcGatewayMux: grpcGatewayMux,
	}

	return apiserver, nil
}

func (a *APIServer) Run(stopCh <-chan struct{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, rfn := range a.registerFns {
		if err := rfn(ctx, a.grpcServer, a.grpcGatewayMux, a.grpcClient); err != nil {
			return err
		}
	}

	errCh := make(chan error)
	go func(errCh chan<- error) {
		defer close(errCh)
		errCh <- a.s.ListenAndServeTLS(
			a.config.TLSCertFile, a.config.TLSKeyFile)
	}(errCh)

	select {
	case err := <-errCh:
		return err

	case <-stopCh:
		if err := a.s.Close(); err != nil {
			return err
		}
		// wait till closed
		for range errCh {
		}
		return nil
	}
}
