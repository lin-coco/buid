package buid

import (
	"errors"
	"log"
	"math/rand"
	"strconv"
	"time"
)

// 一些常量，默认化的配置
const (
	bitLen    = uint(64) // 总占位数 + 1
	maxGetNum = 10000000 // 单次最大获取数，不超过一千万

	epoch = int64(1577808000000) // 设置起始时间(时间戳/毫秒): 2020-01-01 00:00:00，有效期69年

	timestampBits = uint(41) // 时间戳占用位数
	autoIncrBits  = uint(14) // 自增数占用位数
	randomBits    = uint(8)  // 随机数占用位数

	timestampMax = int64(-1 ^ (-1 << timestampBits)) // 时间戳最大值
	autoIncrMax  = int64(-1 ^ (-1 << autoIncrBits))  // 自增数最大值
	randomMax    = int64(-1 ^ (-1 << randomBits))    // 随机数最大值

	autoIncrShift  = randomBits                // 自增数左移位数
	timestampShift = randomBits + autoIncrBits // 时间戳左移位数

	ringBufferSize = uint(1 << autoIncrBits) // 环形数组大小
	randBufferSize = ringBufferSize          // 随机数缓存大小
)

var defaultConfig = &Config{
	Epoch:          epoch,
	Bits:           [3]uint{timestampBits, autoIncrBits, randomBits},
	MaxGetNum:      maxGetNum,
	epoch:          epoch,
	timestampBits:  timestampBits,
	autoIncrBits:   autoIncrBits,
	randomBits:     randomBits,
	timestampMax:   timestampMax,
	autoIncrMax:    autoIncrMax,
	randomMax:      randomMax,
	autoIncrShift:  autoIncrShift,
	timestampShift: timestampShift,
	ringBufferSize: ringBufferSize,
	randBufferSize: randBufferSize,
}

type (
	BUID struct {
		uidRingBuffer    chan int64    // 缓存生成的id
		randBuffer       chan int64    // 缓存随机数
		curUsedTimestamp int64         // 当前uid计算时间戳
		semaphore        chan struct{} // 信号量，通知随机数生产者执行
		exitGen          chan struct{} // 退出信号量
		exitRand         chan struct{} // 退出信号量
		config           *Config       // buid 配置
	}
	Config struct {
		// 设置起始时间(时间戳/毫秒): 默认 2020-01-01 00:00:00
		Epoch int64
		// 0.时间戳占用位数 1.自增数占用位数 2.随机数占用位数 总和 = 63
		// bits[0] >= 35 && bits[0] <= 42 [1,139]年
		// bits[1] >= 10  1024
		// bits[2] >= 8   256
		// 默认 41,14,8
		Bits [3]uint
		// 单次最大获取数，默认一千万
		MaxGetNum int

		epoch          int64 // 起始时间
		timestampBits  uint  // 时间戳占用位数
		autoIncrBits   uint  // 自增数占用位数
		randomBits     uint  // 随机数占用位数
		timestampMax   int64 // 时间戳最大值
		autoIncrMax    int64 // 自增数最大值
		randomMax      int64 // 随机数最大值
		autoIncrShift  uint  // 自增数左移位数
		timestampShift uint  // 时间戳左移位数
		ringBufferSize uint  // 环形数组大小
		randBufferSize uint  // 随机数缓存大小，原则上等于环形数组大小
	}
)

func NewDefaultBUID() (*BUID, error) {
	c := defaultConfig // 复制一份
	if err := validate(c); err != nil {
		return nil, err
	}
	return &BUID{
		uidRingBuffer: make(chan int64, ringBufferSize),
		randBuffer:    make(chan int64, randBufferSize),
		semaphore:     make(chan struct{}),
		exitGen:       make(chan struct{}, 1),
		exitRand:      make(chan struct{}, 1),
		config:        c,
	}, nil
}

func NewBUID(c *Config) (*BUID, error) {
	if c.Epoch == 0 {
		c.Epoch = epoch
	}
	if c.Bits[0] == 0 && c.Bits[1] == 0 && c.Bits[2] == 0 {
		c.Bits = [3]uint{timestampBits, autoIncrBits, randomBits}
	}
	if c.MaxGetNum == 0 {
		c.MaxGetNum = maxGetNum
	}
	if err := validate(c); err != nil {
		return nil, err
	}
	fillConfig(c)
	return &BUID{
		uidRingBuffer: make(chan int64, c.ringBufferSize),
		randBuffer:    make(chan int64, c.randBufferSize),
		semaphore:     make(chan struct{}),
		exitGen:       make(chan struct{}, 1),
		exitRand:      make(chan struct{}, 1),
		config:        c,
	}, nil
}

// validate 校验配置项是否正确
func validate(c *Config) error {
	milli := time.Now().UnixMilli()
	if c.Epoch > milli {
		return errors.New("invalid epoch config, must be less than the current time")
	}
	if c.Bits[0]+c.Bits[1]+c.Bits[2] != bitLen-1 {
		return errors.New("invalid bit config, length must be equal to 63")
	}
	if c.MaxGetNum < 0 {
		return errors.New("invalid maxGetNum config, value must be a positive integer")
	}
	if milli-c.Epoch > int64(-1^(-1<<c.Bits[0])) {
		return errors.New("exceeding the current configuration usage period")
	}
	return nil
}

func fillConfig(c *Config) {
	c.epoch = c.Epoch
	c.timestampBits = c.Bits[0]
	c.autoIncrBits = c.Bits[1]
	c.randomBits = c.Bits[2]
	c.timestampMax = int64(-1 ^ (-1 << c.timestampBits))
	c.autoIncrMax = int64(-1 ^ (-1 << c.autoIncrBits))
	c.randomMax = int64(-1 ^ (-1 << c.randomBits))
	c.autoIncrShift = c.randomBits
	c.timestampShift = c.randomBits + c.autoIncrBits
	c.ringBufferSize = uint(1 << c.autoIncrBits)
	c.randBufferSize = c.ringBufferSize
}

// Start 异步执行
func (b *BUID) Start() {
	go b.runGenerate()
}

// runGenerate 同步执行
func (b *BUID) runGenerate() {
	go b.randomProducer()
LOOP:
	for {
		bNow := time.Now().UnixMilli() - b.config.epoch
		if bNow >= b.config.timestampMax {
			log.Panicf("epoch must be between 0 and %d", b.config.timestampMax)
		}
		if b.curUsedTimestamp > bNow {
			// 发生了时钟回拨 | 在同一毫秒
			b.curUsedTimestamp = b.curUsedTimestamp + int64(rand.Intn(1000))
			log.Printf("suspected clock callback, will use the previous timestamp plus a random number of milliseconds in one second: %d", b.curUsedTimestamp)
		} else {
			// 没有发生时钟回拨
			b.curUsedTimestamp = bNow
		}
		// TODO 保存上一次使用时间戳，持久化功能（文件、etcd）
		select {
		case <-b.exitGen:
			close(b.uidRingBuffer)
			break LOOP
		default:
			// 生成随机数
			b.wakeupRandom()
			autoIncr := int64(0)
			for {
				if autoIncr >= b.config.autoIncrMax {
					// 这一轮增长结束，开始下一轮
					break
				}
				b.uidRingBuffer <- (b.curUsedTimestamp << b.config.timestampShift) | (autoIncr << b.config.autoIncrShift) | <-b.randBuffer
				autoIncr++
			}
		}
	}
}

func (b *BUID) wakeupRandom() {
	b.semaphore <- struct{}{}
}

func (b *BUID) randomProducer() {
	for {
		select {
		case <-b.semaphore:
			r := rand.New(rand.NewSource(b.curUsedTimestamp))
			for i := int64(0); i < b.config.autoIncrMax; i++ {
				r.Int()
				b.randBuffer <- r.Int63n(b.config.randomMax)
			}
		case <-b.exitRand:
			close(b.randBuffer)
			return
		}
	}
}

func (b *BUID) Exit() {
	b.exitGen <- struct{}{}
	for {
		_, ok := <-b.uidRingBuffer
		if !ok {
			break
		}
	}
	b.exitRand <- struct{}{}
	close(b.semaphore)
	close(b.exitGen)
	close(b.exitRand)
}

func (b *BUID) GetUID() int64 {
	return <-b.uidRingBuffer
}

func (b *BUID) GetUIDString() string {
	return strconv.FormatInt(<-b.uidRingBuffer, 10)
}

func (b *BUID) Gets(num int) []int64 {
	ids := make([]int64, num)
	for i := 0; i < num; i++ {
		ids[i] = <-b.uidRingBuffer
	}
	return ids
}

func (b *BUID) GetStrings(num int) []string {
	ids := make([]string, num)
	for i := 0; i < num; i++ {
		ids[i] = strconv.FormatInt(<-b.uidRingBuffer, 10)
	}
	return ids
}

/*
使用方法
1. 作为库使用
func main() {
	buid := internal.NewDefaultBUID()
	buid.StartGenerate()

	now := time.Now()
	_ = buid.Gets(1000000)
	fmt.Println(time.Since(now).Milliseconds())
	buid.Exit()
}
2. 作为服务使用
...
*/
