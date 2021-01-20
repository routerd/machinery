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
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	machineryv1 "routerd.net/machinery/api/v1"
	apiserverv1 "routerd.net/machinery/apiserver/v1"
)

func TestAPIServer(t *testing.T) {
	const address = "localhost:8889"

	s, err := NewServer(APIServerConfig{
		Address:     "localhost:8889",
		TLSCertFile: "cert.pem",
		TLSKeyFile:  "key.pem",
	}, func(ctx context.Context, grpcServer *grpc.Server, grpcGatewayMux *gwruntime.ServeMux, grpcClient *grpc.ClientConn) error {
		healthService := &apiserverv1.HealthServiceServer{}
		machineryv1.RegisterHealthServiceServer(grpcServer, healthService)
		return machineryv1.RegisterHealthServiceHandler(ctx, grpcGatewayMux, grpcClient)
	})
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	stopCh := make(chan struct{})
	go func() {
		err := s.Run(stopCh)
		require.NoError(t, err)

		wg.Done()
	}()

	time.Sleep(1 * time.Second)

	// Load client cert
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	require.NoError(t, err)

	// Load CA cert
	caCert, err := ioutil.ReadFile("cert.pem")
	require.NoError(t, err)

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
	tlsConfig.BuildNameToCertificate()

	t.Run("HTTP", func(t *testing.T) {
		transport := &http.Transport{TLSClientConfig: tlsConfig}
		client := &http.Client{Transport: transport}

		resp, err := client.Get("https://" + address)
		require.NoError(t, err)
		defer resp.Body.Close()
	})

	t.Run("grpc", func(t *testing.T) {
		client, err := grpc.Dial(address, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
		require.NoError(t, err)

		ctx := context.Background()

		healthClient := machineryv1.NewHealthServiceClient(client)
		resp, err := healthClient.Health(ctx, &machineryv1.HealthRequest{})
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	close(stopCh)
	wg.Wait()
}
