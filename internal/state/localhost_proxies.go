package state

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/eleven-sh/agent/internal/env"
	"github.com/eleven-sh/agent/internal/network"
)

type localhostListenerID string

type localhostListener struct {
	listeningPort string
	listeningAddr string
}

type localhostListeners map[localhostListenerID]*localhostListener

type localhostProxy struct {
	listeningPort string
	targetAddr    string
	netListener   net.Listener
	doneChan      chan struct{}
}

var localhostProxies = map[localhostListenerID]*localhostProxy{}

func ReconcileLocalhostProxies(servedPorts env.ConfigServedPorts) error {
	tcpConns, err := network.GetOpenedTCPConns()

	if err != nil {
		return err
	}

	listeners := localhostListeners{}

	for _, conn := range tcpConns {
		if conn.St != uint64(network.TCPConnStatusListening) {
			continue
		}

		if !conn.LocalAddr.IsLoopback() {
			continue
		}

		listeningPortS := fmt.Sprintf("%d", conn.LocalPort)

		if _, isServed := servedPorts[env.ConfigServedPort(listeningPortS)]; !isServed {
			continue
		}

		listeningAddr := conn.LocalAddr.String()

		if conn.LocalAddr.To4() == nil { // IPv6
			listeningAddr = "[" + listeningAddr + "]"
		}

		listenerAddrAndPort := net.JoinHostPort(
			listeningAddr,
			listeningPortS,
		)

		listeners[localhostListenerID(listenerAddrAndPort)] = &localhostListener{
			listeningAddr: listeningAddr,
			listeningPort: listeningPortS,
		}
	}

	reconcileLocalhostProxiesState(listeners)

	return nil
}

func reconcileLocalhostProxiesState(listeners localhostListeners) {
	for listenerID, proxy := range localhostProxies {
		if _, listenerExists := listeners[listenerID]; listenerExists {
			continue
		}

		close(proxy.doneChan)
		delete(localhostProxies, listenerID)
	}

	for listenerID, listener := range listeners {
		if _, proxyExists := localhostProxies[listenerID]; proxyExists {
			continue
		}

		proxy := &localhostProxy{
			targetAddr:    listener.listeningAddr,
			listeningPort: listener.listeningPort,
			doneChan:      make(chan struct{}),
		}

		proxyNetListener, err := startLocalhostProxy(proxy)

		if err != nil {
			log.Printf(
				"[Localhost proxies] Error when starting proxy for %s: %v",
				net.JoinHostPort(proxy.targetAddr, proxy.listeningPort),
				err,
			)

			continue
		}

		proxy.netListener = proxyNetListener

		localhostProxies[listenerID] = proxy

		go handleLocalhostProxyConn(proxy)
	}
}

func startLocalhostProxy(proxy *localhostProxy) (net.Listener, error) {
	ip, err := network.GetOutboundIP()

	if err != nil {
		return nil, err
	}

	return net.Listen(
		"tcp",
		net.JoinHostPort(ip.String(), proxy.listeningPort),
	)
}

func handleLocalhostProxyConn(proxy *localhostProxy) {
	go func() {
		<-proxy.doneChan

		if err := proxy.netListener.Close(); err != nil {
			log.Printf(
				"[Localhost proxies] Error when closing proxy for %s: %v",
				net.JoinHostPort(proxy.targetAddr, proxy.listeningPort),
				err,
			)
		}
	}()

	for {
		proxyConn, err := proxy.netListener.Accept()

		if err != nil {
			select {
			case <-proxy.doneChan:
				return
			default:
				log.Printf(
					"[Localhost proxies] Error when accepting connection on proxy for %s: %v",
					net.JoinHostPort(proxy.targetAddr, proxy.listeningPort),
					err,
				)

				continue
			}
		}

		localConn, err := connectToLocalhostTarget(proxy)

		if err != nil {
			log.Printf(
				"[Localhost proxies] Error when connecting to %s: %v",
				net.JoinHostPort(proxy.targetAddr, proxy.listeningPort),
				err,
			)

			if err := proxyConn.Close(); err != nil {
				log.Printf(
					"[Localhost proxies] Error when closing proxy connection: %v",
					err,
				)
			}

			continue
		}

		go forwardProxyConnToLocalhost(
			proxyConn,
			localConn,
		)
	}
}

func connectToLocalhostTarget(proxy *localhostProxy) (net.Conn, error) {
	return net.Dial(
		"tcp",
		net.JoinHostPort(proxy.targetAddr, proxy.listeningPort),
	)
}

func forwardProxyConnToLocalhost(
	proxyConn net.Conn,
	localConn net.Conn,
) error {

	defer func() {
		proxyConn.Close()
		localConn.Close()
	}()

	proxyConnChan := make(chan error, 1)
	localConnChan := make(chan error, 1)

	// Forward local -> proxy
	go func() {
		_, err := io.Copy(proxyConn, localConn)
		localConnChan <- err
	}()

	// Forward proxy -> local
	go func() {
		_, err := io.Copy(localConn, proxyConn)
		proxyConnChan <- err
	}()

	select {
	case proxyConnErr := <-proxyConnChan:
		if proxyConnErr != nil {
			log.Printf(
				"[Localhost proxies] Error during proxy connection forwarding: %v",
				proxyConnErr,
			)
		}
		return proxyConnErr
	case localConnErr := <-localConnChan:
		if localConnErr != nil {
			log.Printf(
				"[Localhost proxies] Error during local connection forwarding: %v",
				localConnErr,
			)
		}
		return localConnErr
	}
}
