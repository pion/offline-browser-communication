package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"github.com/pion/ice/v2"
	"github.com/pion/webrtc/v3"
)

const remoteDescriptionTemplate = `v=0
o=- 6920920643910646739 2 IN IP4 127.0.0.1
s=-
t=0 0
a=group:BUNDLE 0
a=msid-semantic: WMS
m=application 9 UDP/DTLS/SCTP webrtc-datachannel
c=IN IP4 0.0.0.0
a=ice-ufrag:V6j+
a=ice-pwd:OEKutPgoHVk/99FfqPOf444w
a=fingerprint:sha-256 invalidFingerprint
a=setup:actpass
a=mid:0
a=sctp-port:5000
`

func main() {
	s := webrtc.SettingEngine{}

	// Generate mDNS Candidates and set a static local hostname
	s.SetICEMulticastDNSMode(ice.MulticastDNSModeQueryAndGather)
	s.SetMulticastDNSHostName("offline-browser-communication.local")

	// Set a small number of pre-determined ports we listen for ICE traffic on
	panicIfErr(s.SetEphemeralUDPPortRange(5000, 5005))

	// Disable DTLS Certificate Verification. Currently we aren't able to use stored certificate in the browser
	s.DisableCertificateFingerprintVerification(true)

	// Set static ICE Credentials
	s.SetICECredentials("fKVhbscsMWDGAnBg", "xGjQkAvKIVkBeVTGWcvCQtnVAeapczwa")

	// Create a new PeerConnection, this listens for all incoming DataChannel messages
	api := webrtc.NewAPI(webrtc.WithSettingEngine(s))
	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{
		Certificates: loadCertificate(),
	})
	panicIfErr(err)

	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnOpen(func() {
			fmt.Printf("DataChannel %s has opened \n", d.Label())
		})

		d.OnMessage(func(m webrtc.DataChannelMessage) {
			fmt.Printf("%s \n", m.Data)
		})
	})

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	panicIfErr(peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  remoteDescriptionTemplate,
	}))

	answer, err := peerConnection.CreateAnswer(nil)
	panicIfErr(err)
	panicIfErr(peerConnection.SetLocalDescription(answer))

	fmt.Println("Ready to connect, please load https://jsfiddle.net/nah7qvkj/")
	select {}
}

// If you change this certificate you MUST update the fingerprint in the jsfiddle
func loadCertificate() []webrtc.Certificate {
	certFile, err := ioutil.ReadFile("cert.pem")
	panicIfErr(err)

	keyFile, err := ioutil.ReadFile("key.pem")
	panicIfErr(err)

	certPem, _ := pem.Decode(certFile)
	keyPem, _ := pem.Decode(keyFile)

	cert, err := x509.ParseCertificate(certPem.Bytes)
	panicIfErr(err)

	privateKey, err := x509.ParsePKCS8PrivateKey(keyPem.Bytes)
	panicIfErr(err)

	return []webrtc.Certificate{webrtc.CertificateFromX509(privateKey, cert)}
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
