package generator

import (
	"encoding/base64"
	"fmt"
	"github.com/blademainer/commons/pkg/sign"
	"github.com/blademainer/commons/pkg/util"
	"hash/fnv"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

// Generator generate order ids
type Generator struct {
	ClusterID           string
	MachineID           string
	ClusterAndMachineID string
	Concurrency         int
	maxIndex            int32
	indexWidth          int
	index               int32
	byteLength          int
	debug               bool
}

// zeroByte zero byte
const zeroByte = byte('0')

// New create generator
func New(clusterID string, concurrency int) *Generator {
	if concurrency <= 0 {
		panic("concurrency must greater than 0")
	}
	g := &Generator{}
	id := fmt.Sprint(getIdentityID())
	g.MachineID = id
	g.ClusterID = clusterID
	g.ClusterAndMachineID = fmt.Sprintf("%s%s", clusterID, id)
	g.Concurrency = concurrency
	g.maxIndex = int32(concurrency)
	for g.indexWidth = 0; concurrency > 0; g.indexWidth++ {
		concurrency = concurrency / 10
	}
	dateStr := dateStr()
	g.byteLength = len(dateStr) + len(g.ClusterID) + len(g.MachineID) + g.indexWidth
	return g
}

// Debug debug generator
func (g *Generator) Debug() {
	g.debug = true
}

func getIdentityID() uint32 {
	if name, err := os.Hostname(); err == nil {
		h := fnv.New32()
		_, err := h.Write([]byte(name))
		if err != nil {
			panic(err.Error())
		}
		sum32 := h.Sum32()
		return sum32
	}
	macs := util.GetMacAddrs()

	mac := ""
	if len(macs) == 0 {
		rsaGenerator, err := sign.NewRsa2048Generator()
		if err != nil {
			panic(err.Error())
		}
		mac, err = rsaGenerator.GeneratePemPublicKey()
		if err != nil {
			mac = util.RandString(64)
		}
	}
	h := fnv.New32()
	_, err := h.Write([]byte(mac))
	if err != nil {
		panic(err.Error())
	}
	sum32 := h.Sum32()
	return sum32
}

func dateStr() []byte {
	// b := strings.Builder{}
	now := time.Now()
	// year := now.Year()
	// month := int(now.Month())
	// day := now.Day()
	year, month, day := now.Date()
	hour := now.Hour()
	minute := now.Minute()
	second := now.Second()
	date := int64(year)
	date = date*int64(100) + int64(month)
	date = date*int64(100) + int64(day)
	date = date*int64(100) + int64(hour)
	date = date*int64(100) + int64(minute)
	date = date*int64(100) + int64(second)
	// date := now.Format(TIME_LAYOUT)
	millSeconds := now.Nanosecond() / 1000000
	date = date*int64(1000) + int64(millSeconds)
	bts := make([]byte, 0, 17)
	bts = strconv.AppendInt(bts, date, 10)
	// b.WriteString(strconv.FormatInt(date, 10))
	// b.WriteString(strconv.Itoa(nanosecond))
	return bts
}

// GenerateIndex generator index
func (g *Generator) GenerateIndex() []byte {
	rs := make([]byte, g.indexWidth)
	index := atomic.AddInt32(&g.index, 1) % g.maxIndex
	if index < 0 {
		index = -index
		g.index = -g.index
	}
	wi := strconv.Itoa(int(index))

	// wi := strconv.AppendInt([]byte{}, int64(index), 10)
	start := g.indexWidth - len(wi)
	copy(rs[start:g.indexWidth], wi)
	for i := 0; i < start; i++ {
		rs[i] = zeroByte
	}
	return rs
}

// GenerateID generate id
func (g *Generator) GenerateID() string {

	rs := make([]byte, g.byteLength)
	dateStr := dateStr()
	i := 0
	c := len(dateStr)
	copy(rs[i:c], dateStr)

	i = c
	c += len(g.ClusterAndMachineID)
	copy(rs[i:c], g.ClusterAndMachineID)

	index := g.GenerateIndex()
	i = c
	c += len(index)
	copy(rs[i:c], index)
	if g.debug {
		fmt.Printf(
			"date[%s] + clusterID[%s] + machineID[%s] + index[%s] + byteLen: %v + cap: %v"+
				", hex: %v\n", dateStr,
			g.ClusterID, g.MachineID, index, g.byteLength, c,
			base64.RawStdEncoding.EncodeToString(rs),
		)

	}
	return string(rs)
}
