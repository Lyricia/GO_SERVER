package Protocol

const (
	WORLD_WIDTH  = 8
	WORLD_HEIGHT = 8

	SERVER_PORT = 8000
	SINGLE_ID   = 1

	CS_UP    uint8 = 1
	CS_DOWN  uint8 = 2
	CS_LEFT  uint8 = 3
	CS_RIGHT uint8 = 4

	SC_LOGIN_OK      uint8 = 1
	SC_PUT_PLAYER    uint8 = 2
	SC_REMOVE_PLAYER uint8 = 3
	SC_POS           uint8 = 4
)

// Go 역시 Padding이 있고 문제는 #pragma pack이 없음
// 구조체 짤때 미리 어느정도 순서를 어레인지 해야할거같음
// https://johngrib.github.io/wiki/golang-struct-padding/ 참고

// 그냥 byte buffer에 binary write로 씌워버리니까 해결

type Packet_SC_POS struct {
	Size        uint8
	Packet_type uint8
	Id          uint32
	X, Y        uint8
}

type Packet_SC_Remove_Player struct {
	Size        uint8
	Packet_type uint8
	Id          uint32
}

type Packet_SC_Login_OK struct {
	Size        uint8
	Packet_type uint8
	Id          uint32
}

type Packet_SC_Put_Player struct {
	Size        uint8
	Packet_type uint8
	Id          uint32
	X, Y        uint16
}

type CS_Packet_Move struct {
	Size        uint8
	Packet_type uint8
}
