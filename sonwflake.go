package twig

import (
	"errors"
	"math/rand"
	"net"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type IdGenerator interface {
	NextID() uint64
}

const (
	BitLenTime      = 39                               // bit length of time
	BitLenSequence  = 8                                // bit length of sequence number
	BitLenMachineID = 63 - BitLenTime - BitLenSequence // bit length of machine id
)

type Sonwflake struct {
	mutex       *sync.Mutex
	startTime   int64
	elapsedTime int64
	sequence    uint16
	machineID   uint16
}

func NewSonwflake() *Sonwflake {
	sf := new(Sonwflake)
	sf.mutex = new(sync.Mutex)
	sf.sequence = uint16(1<<BitLenSequence - 1)
	sf.startTime = toSonwflakeTime(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC))

	var err error
	if sf.machineID, err = lower16BitPrivateIP(); err != nil {
		sf.machineID = uint16(rand.Uint32())
	}

	return sf
}

func (sf *Sonwflake) NextID() uint64 {
	const maskSequence = uint16(1<<BitLenSequence - 1)

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

	return mustID(sf.toID())
}

func mustID(i uint64, e error) uint64 {
	return i
}

func (sf *Sonwflake) toID() (uint64, error) {
	if sf.elapsedTime >= 1<<BitLenTime {
		return 0, errors.New("over the time limit")
	}

	return uint64(sf.elapsedTime)<<(BitLenSequence+BitLenMachineID) |
		uint64(sf.sequence)<<BitLenMachineID |
		uint64(sf.machineID), nil
}

const sonwflakeTimeUnit = 1e7 // nsec, i.e. 10 msec

func toSonwflakeTime(t time.Time) int64 {
	return t.UTC().UnixNano() / sonwflakeTimeUnit
}

func currentElapsedTime(startTime int64) int64 {
	return toSonwflakeTime(time.Now()) - startTime
}

func sleepTime(overtime int64) time.Duration {
	return time.Duration(overtime)*10*time.Millisecond -
		time.Duration(time.Now().UTC().UnixNano()%sonwflakeTimeUnit)*time.Nanosecond
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

type IdGen struct {
	IdGenerator
}

const idGenID = "_twig_snowflake_id_gen_"

func (id *IdGen) Name() string {
	return idGenID
}

func (id *IdGen) ID() string {
	return idGenID
}

func (id *IdGen) Type() string {
	return "plugin"
}

func GetIdGenerator(c Ctx) IdGenerator {
	p, _ := GetPlugin(idGenID, c)
	return p.(IdGenerator)
}
