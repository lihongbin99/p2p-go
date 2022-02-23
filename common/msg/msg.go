package msg

const (
	_ byte = iota
	ERROR

	Ignore

	Ping
	Pong

	IdRequest
	IdResponse

	UDPRegisterRequest
	UDPRegisterResponse
	UDPAccessRequest
	UDPAccessResponse
	UDPAccessCloseResponse
	UDPAccessFlushResponse

	UDPNewConnectRequest
	UDPNewConnectResponse
	UDPNewConnectResultRequest
	UDPNewConnectResultResponse

	TCPRegisterRequest
	TCPRegisterResponse
	TCPAccessRequest
	TCPAccessResponse
	TCPAccessCloseResponse
	TCPAccessFlushResponse

	TCPNewConnectRequest
	TCPNewConnectResponse
	TCPNewConnectResultRequest
	TCPNewConnectResultResponse

	TCPTransferRequest
	TCPTransferResponse
	UDPTransferRequest
	UDPTransferResponse

	TCPTunRegisterRequest
	TCPTunRegisterResponse
	TCPTunNewConnectRequest
	TCPTunNewConnectResponse
)

type Message interface {
	toMessage(srcBuf []byte) Message
	ToByteBuf() []byte
}

var (
	types = map[byte]Message{
		ERROR: &ErrorMessage{},

		Ignore: &IgnoreMessage{},

		Ping: &PingMessage{},
		Pong: &PongMessage{},

		IdRequest:  &IdRequestMessage{},
		IdResponse: &IdResponseMessage{},

		UDPRegisterRequest:     &UDPRegisterRequestMessage{},
		UDPRegisterResponse:    &UDPRegisterResponseMessage{},
		UDPAccessRequest:       &UDPAccessRequestMessage{},
		UDPAccessResponse:      &UDPAccessResponseMessage{},
		UDPAccessCloseResponse: &UDPAccessCloseResponseMessage{},
		UDPAccessFlushResponse: &UDPAccessFlushResponseMessage{},

		UDPNewConnectRequest:        &UDPNewConnectRequestMessage{},
		UDPNewConnectResponse:       &UDPNewConnectResponseMessage{},
		UDPNewConnectResultRequest:  &UDPNewConnectResultRequestMessage{},
		UDPNewConnectResultResponse: &UDPNewConnectResultResponseMessage{},

		TCPRegisterRequest:     &TCPRegisterRequestMessage{},
		TCPRegisterResponse:    &TCPRegisterResponseMessage{},
		TCPAccessRequest:       &TCPAccessRequestMessage{},
		TCPAccessResponse:      &TCPAccessResponseMessage{},
		TCPAccessCloseResponse: &TCPAccessCloseResponseMessage{},
		TCPAccessFlushResponse: &TCPAccessFlushResponseMessage{},

		TCPNewConnectRequest:        &TCPNewConnectRequestMessage{},
		TCPNewConnectResponse:       &TCPNewConnectResponseMessage{},
		TCPNewConnectResultRequest:  &TCPNewConnectResultRequestMessage{},
		TCPNewConnectResultResponse: &TCPNewConnectResultResponseMessage{},

		TCPTransferRequest:  &TCPTransferRequestMessage{},
		TCPTransferResponse: &TCPTransferResponseMessage{},
		UDPTransferRequest:  &UDPTransferRequestMessage{},
		UDPTransferResponse: &UDPTransferResponseMessage{},

		TCPTunRegisterRequest:    &TCPTunRegisterRequestMessage{},
		TCPTunRegisterResponse:   &TCPTunRegisterResponseMessage{},
		TCPTunNewConnectRequest:  &TCPTunNewConnectRequestMessage{},
		TCPTunNewConnectResponse: &TCPTunNewConnectResponseMessage{},
	}
	errorMessage = &ErrorMessage{}
	layout       = "06:01:02 15:04:05"
)

func Read(buf []byte) (message Message) {
	if len(buf) > 0 {
		if ref, ok := types[buf[0]]; ok {
			if message = ref.toMessage(buf); message != nil {
				return
			}
		}
	}
	message = errorMessage.toMessage(buf)
	return
}
