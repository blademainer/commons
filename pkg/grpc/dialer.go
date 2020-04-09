package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"net"
	"net/url"
	"strings"
)

// ParseDialTarget returns the network and address to pass to dialer
func ParseDialTarget(target string) (net string, addr string) {
	net = "tcp"

	m1 := strings.Index(target, ":")
	m2 := strings.Index(target, ":/")

	// handle unix:addr which will fail with url.Parse
	if m1 >= 0 && m2 < 0 {
		if n := target[0:m1]; n == "unix" {
			net = n
			addr = target[m1+1:]
			return net, addr
		}
	}
	if m2 >= 0 {
		t, err := url.Parse(target)
		if err != nil {
			return net, target
		}
		scheme := t.Scheme
		addr = t.Path
		if scheme == "unix" {
			net = scheme
			if addr == "" {
				addr = t.Host
			}
			return net, addr
		}
	}

	return net, target
}

func DialUrl(ctx context.Context, addr url.URL, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials
	if addr.Scheme == "https" {
		var err error
		creds, err = ClientTransportCredentials(true, "", "", "")
		if err != nil {
			return nil, err
		}
	}
	network, host := ParseDialTarget(addr.String())
	host = addr.Host
	return Dial(ctx, network, host, creds, opts...)
}

func Dial(ctx context.Context, network, address string, creds credentials.TransportCredentials, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	dialer := func(ctx context.Context, address string) (net.Conn, error) {
		conn, err := (&net.Dialer{}).DialContext(ctx, network, address)
		if err != nil {
			return nil, err
		}
		if creds != nil {
			conn, _, err = creds.ClientHandshake(ctx, address, conn)
			if err != nil {
				return nil, err
			}
		}
		return conn, nil
	}

	opts = append(opts,
		grpc.FailOnNonTempDialError(true),
		grpc.WithInsecure(), // we are handling TLS, so tell grpc not to
		grpc.WithContextDialer(dialer),
	)

	// Set up a connection to the server.
	conn, err := grpc.DialContext(ctx, address, opts...)
	return conn, err
}

// ClientTransportCredentials builds transport credentials for a gRPC client using the
// given properties. If cacertFile is blank, only standard trusted certs are used to
// verify the server certs. If clientCertFile is blank, the client will not use a client
// certificate. If clientCertFile is not blank then clientKeyFile must not be blank.
func ClientTransportCredentials(insecureSkipVerify bool, cacertFile, clientCertFile, clientKeyFile string) (credentials.TransportCredentials, error) {
	var tlsConf tls.Config

	if clientCertFile != "" {
		// Load the client certificates from disk
		certificate, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
		if err != nil {
			return nil, fmt.Errorf("could not load client key pair: %v", err)
		}
		tlsConf.Certificates = []tls.Certificate{certificate}
	}

	if insecureSkipVerify {
		tlsConf.InsecureSkipVerify = true
	} else if cacertFile != "" {
		// Create a certificate pool from the certificate authority
		certPool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(cacertFile)
		if err != nil {
			return nil, fmt.Errorf("could not read ca certificate: %v", err)
		}

		// Append the certificates from the CA
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			return nil, errors.New("failed to append ca certs")
		}

		tlsConf.RootCAs = certPool
	}

	return credentials.NewTLS(&tlsConf), nil
}
