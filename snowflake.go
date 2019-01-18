package twig

import (
	"errors"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type IdGenerator interface {
	NextID() string
}

const (
	bitLenTime      = 39                               // bit length of time
	bitLenSequence  = 8                                // bit length of sequence number
	bitLenMachineID = 63 - bitLenTime - bitLenSequence // bit length of machine id
)

type Snowflake struct {
	mutex       *sync.Mutex
	startTime   int64
	elapsedTime int64
	sequence    uint16
	machineID   uint16
}

type NodeIdFunc func() uint16

func nodeid() uint16 {
	if id, err := lower16BitPrivateIP(); err == nil {
		return id
	}
	return uint16(rand.Uint32())
}

func NewSnowflake(id NodeIdFunc) *Snowflake {
	sf := new(Snowflake)
	sf.mutex = new(sync.Mutex)
	sf.sequence = uint16(1<<bitLenSequence - 1)
	sf.startTime = toSnowflakeTime(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC))
	sf.machineID = id()
	return sf
}

func (sf *Snowflake) NextID() (uint64, error) {
	const maskSequence = uint16(1<<bitLenSequence - 1)

	sf.mutex.Lock()
	defer sf.mutex.Unlock()

	current := currentElapsedTime(sf.startTime)
	if sf.elapsedTime < current {
		sf.elapsedTime = current
		sf.sequence = 0
	} else { // sf.elapsedTime >= current
		sf.sequence = (sf.sequence + 1) & maskSequence
		if sf.sequence == 0 {
			sf.elapsedTime++
			overtime := sf.elapsedTime - current
			time.Sleep(sleepTime((overtime)))
		}
	}

	return sf.toID()
}

func (sf *Snowflake) toID() (uint64, error) {
	if sf.elapsedTime >= 1<<bitLenTime {
		return 0, errors.New("over the time limit")
	}

	return uint64(sf.elapsedTime)<<(bitLenSequence+bitLenMachineID) |
		uint64(sf.sequence)<<bitLenMachineID |
		uint64(sf.machineID), nil
}

const snowflakeTimeUnit = 1e7 // nsec, i.e. 10 msec

func toSnowflakeTime(t time.Time) int64 {
	return t.UTC().UnixNano() / snowflakeTimeUnit
}

func currentElapsedTime(startTime int64) int64 {
	return toSnowflakeTime(time.Now()) - startTime
}

func sleepTime(overtime int64) time.Duration {
	return time.Duration(overtime)*10*time.Millisecond -
		time.Duration(time.Now().UTC().UnixNano()%snowflakeTimeUnit)*time.Nanosecond
}

func privateIPv4() (net.IP, error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if isPrivateIPv4(ip) {
			return ip, nil
		}
	}
	return nil, errors.New("no private ip address")
}

func isPrivateIPv4(ip net.IP) bool {
	return ip != nil &&
		(ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) || ip[0] == 192 && ip[1] == 168)
}

func lower16BitPrivateIP() (uint16, error) {
	ip, err := privateIPv4()
	if err != nil {
		return 0, err
	}

	return uint16(ip[2])<<8 + uint16(ip[3]), nil
}

func MustNextID(i uint64, e error) uint64 {
	if e == nil {
		return i
	}
	panic(e)
}

type snowflakeIdGen struct {
	sonwflake *Snowflake
}

func (id *snowflakeIdGen) NextID() string {
	i := MustNextID(id.sonwflake.NextID())
	return strconv.FormatUint(i, 32)
}

const snowflakePlugID = "_twig_snowflake_id_gen_"

func (id *snowflakeIdGen) Name() string {
	return snowflakePlugID
}

func (id *snowflakeIdGen) ID() string {
	return snowflakePlugID
}

func (id *snowflakeIdGen) Type() string {
	return "plugin"
}

func GetIdGenerator(c Ctx) IdGenerator {
	p, _ := GetPlugin(snowflakePlugID, c)
	return p.(IdGenerator)
}
